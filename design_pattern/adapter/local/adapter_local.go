package local

import "github.com/msyamsula/portofolio/design_pattern/adapter/fixed"

type LocalAdapter struct {
	logger *fixed.LocalLog
}

func (fd *LocalAdapter) LogInfo(message string) {
	fd.logger.Log(message)

}
func (fd *LocalAdapter) LogError(message string) {
	fd.logger.Log(message)
}
