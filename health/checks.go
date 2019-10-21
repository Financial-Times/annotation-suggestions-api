package health

import (
	"context"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
)

func (service *HealthService) draftContentCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to provide suggestions to editorial for tagging content",
		Name:             "Draft Content Service Health Check",
		PanicGuide:       "https://biz-ops.in.ft.com/System/draft-content-suggestions",
		Severity:         1,
		TechnicalSummary: "Checks whether the health endpoint of draft-content-api returns successful responses",
		Checker:          service.draftContentChecker,
	}
}

func (service *HealthService) suggestionsCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to provide suggestions to editorial for tagging content",
		Name:             "Suggestions Umbrella Service Health Check",
		PanicGuide:       "https://biz-ops.in.ft.com/System/draft-content-suggestions",
		Severity:         1,
		TechnicalSummary: "Checks whether the suggestions umbrella endpoint is accessible and returns responses",
		Checker:          service.suggestionsChecker,
	}
}

func (service *HealthService) draftContentChecker() (string, error) {
	return service.contentAPI.IsGTG(context.Background())
}

func (service *HealthService) suggestionsChecker() (string, error) {
	return service.umbrellaAPI.IsGTG(context.Background())
}
