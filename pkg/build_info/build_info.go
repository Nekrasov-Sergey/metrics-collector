package buildinfo

import (
	"fmt"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

func Print(buildInfo types.BuildInfo) {
	fmt.Println("Build version:", buildInfo.Version)
	fmt.Println("Build date:", buildInfo.Date)
	fmt.Println("Build commit:", buildInfo.Commit)
}
