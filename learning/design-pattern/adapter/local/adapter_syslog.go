package local

import "github.com/msyamsula/portofolio/other-works/design-pattern/adapter/fixed"

type SysLogAdapter struct {
	logger *fixed.SysLog
}

func (fd *SysLogAdapter) LogInfo(message string) {
	fd.logger.Log(1, message)

}
func (fd *SysLogAdapter) LogError(message string) {
	fd.logger.Log(2, message)
}
