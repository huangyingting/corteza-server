package codegen

import (
	"flag"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func Proc() {
	var (
		err error

		watchChanges bool
		beVerbose    bool

		fileList []string
		watcher  *fsnotify.Watcher

		templatesPath = filepath.Join("pkg", "codegen", "assets", "*.tpl")
		templatesSrc  []string

		actionSrcPath = filepath.Join("*", "service", "*_actions.yaml")
		actionSrc     []string
		actionDefs    []*actionsDef

		eventSrcPath = filepath.Join("*", "service", "event", "events.yaml")
		eventSrc     []string
		eventDefs    []*eventsDef

		// workaround because
		// filepath.Join merges "*","*" into "**" instead of "*/*"
		typeSrcPath = filepath.Join("*"+string(filepath.Separator)+"*", "types.yaml")
		typeSrc     []string
		typeDefs    []*typesDef

		restSrcPath = filepath.Join("*", "rest.yaml")
		restSrc     []string
		restDefs    []*restDef

		storeSrcPath = filepath.Join("store", "*.yaml")
		storeSrc     []string
		storeDefs    []*storeDef

		optionSrcPath = filepath.Join("pkg", "options", "*.yaml")
		optionSrc     []string
		optionDefs    []*optionsDef

		tpls    *template.Template
		tplBase = template.New("").
			Funcs(map[string]interface{}{
				"camelCase":       camelCase,
				"export":          export,
				"unexport":        unexport,
				"toggleExport":    toggleExport,
				"toLower":         strings.ToLower,
				"toUpper":         strings.ToUpper,
				"cc2underscore":   cc2underscore,
				"normalizeImport": normalizeImport,
				"comment": func(text string, skip1st bool) string {
					ll := strings.Split(text, "\n")
					s := 0
					out := ""
					if skip1st {
						s = 1
						out = ll[0] + "\n"
					}

					for ; s < len(ll); s++ {
						out += "// " + ll[s] + "\n"
					}

					return out
				},
			}).
			Funcs(sprig.TxtFuncMap())

		output = func(format string, aa ...interface{}) {
			if beVerbose {
				fmt.Fprintf(os.Stdout, format, aa...)
			}
		}

		outputErr = func(err error, format string, aa ...interface{}) bool {
			if err != nil {
				fmt.Fprintf(os.Stdout, format, aa...)
				fmt.Fprintf(os.Stdout, "%v\n", err)
				return true
			}

			return false
		}
	)

	flag.BoolVar(&watchChanges, "w", false, "regenerate code on template or definition change")
	flag.BoolVar(&beVerbose, "v", false, "output loaded definitions, templates and outputs")
	flag.Parse()

	defer func() {
		if watcher != nil {
			watcher.Close()
		}
	}()

	for {
		fileList = make([]string, 0, 100)

		templatesSrc = glob(templatesPath)
		output("loaded %d templates from %s\n", len(templatesSrc), templatesPath)

		actionSrc = glob(actionSrcPath)
		output("loaded %d action definitions from %s\n", len(actionSrc), actionSrcPath)

		eventSrc = glob(eventSrcPath)
		output("loaded %d event definitions from %s\n", len(eventSrc), eventSrcPath)

		typeSrc = glob(typeSrcPath)
		output("loaded %d type definitions from %s\n", len(typeSrc), typeSrcPath)

		restSrc = glob(restSrcPath)
		output("loaded %d rest definitions from %s\n", len(restSrc), restSrcPath)

		storeSrc = glob(storeSrcPath)
		output("loaded %d store definitions from %s\n", len(storeSrc), storeSrcPath)

		optionSrc = glob(optionSrcPath)
		output("loaded %d option defenitions from %s\n", len(optionSrc), optionSrcPath)

		if watchChanges {
			if watcher != nil {
				watcher.Close()
			}

			watcher, err = fsnotify.NewWatcher()
			handleError(err)

			fileList = append(fileList, templatesSrc...)
			fileList = append(fileList, actionSrc...)
			fileList = append(fileList, eventSrc...)
			fileList = append(fileList, typeSrc...)
			fileList = append(fileList, restSrc...)
			fileList = append(fileList, storeSrc...)
			fileList = append(fileList, optionSrc...)

			for _, d := range fileList {
				handleError(watcher.Add(d))
			}
		}

		func() {
			tpls, err = tplBase.ParseFiles(templatesSrc...)
			if outputErr(err, "could not parse templates:\n") {
				return
			}

			if actionDefs, err = procActions(actionSrc...); err == nil {
				err = genActions(tpls, actionDefs...)
			}

			if outputErr(err, "failed to process actions:\n") {
				return
			}

			if eventDefs, err = procEvents(eventSrc...); err == nil {
				err = genEvents(tpls, eventDefs...)
			}

			if outputErr(err, "failed to process events:\n") {
				return
			}

			if typeDefs, err = procTypes(typeSrc...); err == nil {
				err = genTypes(tpls, typeDefs...)
			}

			if outputErr(err, "failed to process types:\n") {
				return
			}

			if restDefs, err = procRest(restSrc...); err == nil {
				err = genRest(tpls, restDefs...)
			}

			if outputErr(err, "failed to process rest:\n") {
				return
			}

			if storeDefs, err = procStore(storeSrc...); err == nil {
				err = genStore(tpls, storeDefs...)
			}

			if outputErr(err, "failed to process store:\n") {
				return
			}

			if optionDefs, err = procOptions(optionSrc...); err == nil {
				err = genOptions(tpls, optionDefs...)
			}

			if outputErr(err, "fail to process options:\n") {
				return
			}

		}()

		if !watchChanges {
			break
		}

		// @todo fix this (without causing too many "too-many-files" issues :)
		output("waiting for changes (if you add a new file, restart codegen manually)\n")

		select {
		case <-watcher.Events:
		case err = <-watcher.Errors:
			handleError(err)
		}
	}
}

func glob(path string) []string {
	src, err := filepath.Glob(path)
	if err != nil {
		handleError(fmt.Errorf("failed to glob %q: %w", path, err))
	}

	return src
}

// Similar to cli.HandleError but without cli pkg dependencies
//
// pkg/cli deps can give us issues when generating go files with invalid code
func handleError(err error) {
	if err == nil {
		return
	}

	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
