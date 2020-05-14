package health

import (
	"time"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
)

const DefaultHealthPath = "/__health"

type HealthService struct {
	config       *HealthConfig
	healthChecks []fthealth.Check
	gtgChecks    []gtg.StatusChecker
	contentAPI   draft.ContentAPI
	umbrellaAPI  suggestions.UmbrellaAPI
}

type HealthConfig struct {
	appSystemCode  string
	appName        string
	appDescription string
}

func NewHealthService(appSystemCode string, appName string,
	appDescription string, contentAPI draft.ContentAPI,
	umbrellaAPI suggestions.UmbrellaAPI) *HealthService {

	hc := &HealthService{
		config: &HealthConfig{
			appSystemCode:  appSystemCode,
			appName:        appName,
			appDescription: appDescription,
		},
		contentAPI:  contentAPI,
		umbrellaAPI: umbrellaAPI,
	}

	hc.healthChecks = []fthealth.Check{hc.draftContentCheck(), hc.suggestionsCheck()}

	draftContentCheck := func() gtg.Status {
		return gtgCheck(hc.draftContentChecker)
	}
	suggestionsCheck := func() gtg.Status {
		return gtgCheck(hc.suggestionsChecker)
	}

	gtgChecks := append(hc.gtgChecks, draftContentCheck, suggestionsCheck)
	hc.gtgChecks = gtgChecks
	return hc
}

func (service *HealthService) Health() fthealth.HC {
	return &fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  service.config.appSystemCode,
			Name:        service.config.appName,
			Description: service.config.appDescription,
			Checks:      service.healthChecks,
		},
		Timeout: 10 * time.Second,
	}
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

func (service *HealthService) GTG() gtg.Status {
	return gtg.FailFastParallelCheck(service.gtgChecks)()
}
