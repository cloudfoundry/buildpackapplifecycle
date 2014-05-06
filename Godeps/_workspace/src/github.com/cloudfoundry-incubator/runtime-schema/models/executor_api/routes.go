package executor_api

import "github.com/tedsuo/router"

const (
	GetContainer        = "GetContainer"
	AllocateContainer   = "AllocateContainer"
	InitializeContainer = "InitializeContainer"
	RunActions          = "RunActions"
	DeleteContainer     = "DeleteContainer"
)

var Routes = router.Routes{
	{Path: "/containers/:guid", Method: "GET", Handler: GetContainer},
	{Path: "/containers", Method: "POST", Handler: AllocateContainer},
	{Path: "/containers/:guid/initialize", Method: "POST", Handler: InitializeContainer},
	{Path: "/containers/:guid/run", Method: "POST", Handler: RunActions},
	{Path: "/containers/:guid", Method: "DELETE", Handler: DeleteContainer},
}
