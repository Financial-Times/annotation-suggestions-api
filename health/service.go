package health

import (
	"context"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/service-status-go/gtg"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

const DefaultHealthPath = "/__health"

type Service struct {
	config       *appConfig
	healthChecks []fthealth.Check
	gtgChecks    []gtg.StatusChecker
	contentAPI   draft.ContentAPI
	umbrellaAPI  suggestions.UmbrellaAPI
	log          *logger.UPPLogger
}

type appConfig struct {
	appSystemCode  string
	appName        string
	appDescription string
}

func NewService(appSystemCode string, appName string,
	appDescription string, contentAPI draft.ContentAPI,
	umbrellaAPI suggestions.UmbrellaAPI, log *logger.UPPLogger) *Service {

	hc := &Service{
		config: &appConfig{
			appSystemCode:  appSystemCode,
			appName:        appName,
			appDescription: appDescription,
		},
		contentAPI:  contentAPI,
		umbrellaAPI: umbrellaAPI,
		log:         log,
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

func (s *Service) Health() fthealth.HC {
	return &fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  s.config.appSystemCode,
			Name:        s.config.appName,
			Description: s.config.appDescription,
			Checks:      s.healthChecks,
		},
		Timeout: 10 * time.Second,
	}
}

func (s *Service) GTG() gtg.Status {
	return gtg.FailFastParallelCheck(s.gtgChecks)()
}

func (s *Service) draftContentCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to provide suggestions to editorial for tagging content",
		Name:             "Draft Content Service Health Check",
		PanicGuide:       "https://runbooks.ftops.tech/draft-content-suggestions",
		Severity:         1,
		TechnicalSummary: "Checks whether the health endpoint of draft-content-api returns successful responses",
		Checker:          s.draftContentChecker,
	}
}

func (s *Service) suggestionsCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to provide suggestions to editorial for tagging content",
		Name:             "Suggestions Umbrella Service Health Check",
		PanicGuide:       "https://runbooks.ftops.tech/draft-content-suggestions",
		Severity:         1,
		TechnicalSummary: "Checks whether the suggestions umbrella endpoint is accessible and returns responses",
		Checker:          s.suggestionsChecker,
	}
}

func (s *Service) draftContentChecker() (string, error) {
	res, err := s.contentAPI.IsGTG(context.Background())
	if err != nil {
		s.log.WithField("healthEndpoint", s.contentAPI.Endpoint()).WithError(err).Error("Draft Content GTG check failed")
	}

	return res, err
}

func (s *Service) suggestionsChecker() (string, error) {
	res, err := s.umbrellaAPI.IsGTG(context.Background())
	if err != nil {
		s.log.WithField("healthEndpoint", s.umbrellaAPI.Endpoint()).WithError(err).Error("UPP Suggestions API GTG check failed")
	}

	return res, err
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}
