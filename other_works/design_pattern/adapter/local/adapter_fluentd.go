package local

import "github.com/msyamsula/portofolio/other_works/design_pattern/adapter/fixed"

type FluendAdapter struct {
	fluend *fixed.FluentD
}

func (fd *FluendAdapter) LogInfo(message string) {
	fd.fluend.Log("Info", message)

}
func (fd *FluendAdapter) LogError(message string) {
	fd.fluend.Log("Error", message)
}
