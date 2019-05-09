package types

import (
	"github.com/crusttech/crust/internal/permissions"
)

const MessagingPermissionResource = permissions.Resource("messaging")
const ChannelPermissionResource = permissions.Resource("messaging:channel:")
const WebhookPermissionResource = permissions.Resource("messaging:webhook:")
