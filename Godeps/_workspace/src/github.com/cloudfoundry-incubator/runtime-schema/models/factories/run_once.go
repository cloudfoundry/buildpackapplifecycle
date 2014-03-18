package factories

import (
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/nu7hatch/gouuid"
)

func GenerateGuid() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic("Failed to generate a GUID.  Craziness.")
	}

	return guid.String()
}

func BuildRunOnceWithRunAction(memoryMB int, diskMB int, script string) *models.RunOnce {
	return &models.RunOnce{
		Guid:     GenerateGuid(),
		MemoryMB: memoryMB,
		DiskMB:   diskMB,
		Actions: []models.ExecutorAction{
			{Action: models.RunAction{
				Script: script,
			}},
		},
	}
}
