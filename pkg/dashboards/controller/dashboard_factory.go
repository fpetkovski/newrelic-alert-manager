package controller

import (
	"fmt"
	"github.com/fpetkovski/newrelic-alert-manager/pkg/apis/newrelic/v1alpha1"
	"github.com/fpetkovski/newrelic-alert-manager/pkg/applications"
	"github.com/fpetkovski/newrelic-alert-manager/pkg/dashboards/domain"
	"github.com/fpetkovski/newrelic-alert-manager/pkg/dashboards/domain/widget"
)

type DashboardFactory struct {
	appRepository *applications.Repository
}

func NewDashboardFactory(appRepository *applications.Repository) *DashboardFactory {
	return &DashboardFactory{
		appRepository: appRepository,
	}
}

func (factory DashboardFactory) NewDashboard(cr *v1alpha1.Dashboard) (*domain.Dashboard, error) {
	widgets, err := factory.newWidgets(cr.Spec.Widgets)
	if err != nil {
		return nil, err
	}

	return &domain.Dashboard{
		DashboardBody: domain.DashboardBody{
			Id:         cr.Status.NewrelicDashboardId,
			Title:      cr.Spec.Title,
			Visibility: "all",
			Editable:   "read_only",
			Metadata: domain.Metadata{
				Version: 1,
			},
			Widgets: widgets,
		},
	}, nil
}

func (factory DashboardFactory) newWidgets(widgets []v1alpha1.Widget) (widget.WidgetList, error) {
	result := make(widget.WidgetList, len(widgets))
	for i, w := range widgets {
		data, err := factory.newData(w.Data)
		if err != nil {
			return nil, err
		}

		result[i] = widget.Widget{
			Visualization: w.Visualization,
			Data:          data,
			Layout:        w.Layout,
			Presentation: widget.Presentation{
				Title: w.Title,
			},
		}
	}

	return result, nil
}

func (factory DashboardFactory) newData(data v1alpha1.Data) (widget.DataList, error) {
	var result widget.DataList

	if data.Nrql != "" && data.ApmMetric != nil {
		return result, fmt.Errorf("you can either set data.nrql or data.apm, but not both")
	}

	if data.Nrql == "" && data.ApmMetric == nil {
		return result, fmt.Errorf("you must set either set data.nrql or data.apm")
	}

	if data.Nrql != "" {
		result[0] = widget.Data{
			Nrql: data.Nrql,
		}
	} else {
		entities, err := factory.getApplicationIds(data.ApmMetric.Entities)
		if err != nil {
			return result, err
		}

		result[0] = widget.Data{
			ApmMetric: &widget.ApmMetric{
				Duration:  data.ApmMetric.Duration,
				EntityIds: entities,
				Metrics:   newMetrics(data.ApmMetric.Metrics),
				Facet:     data.ApmMetric.Facet,
				OrderBy:   data.ApmMetric.OrderBy,
			},
			Nrql: "",
		}
	}

	return result, nil
}

func (factory DashboardFactory) getApplicationIds(entities []string) ([]int, error) {
	var result []int
	for _, item := range entities {
		application, err := factory.appRepository.GetApplicationByName(item)
		if err != nil {
			return nil, err
		}
		if application == nil {
			continue
		}

		result = append(result, application.Id)
	}

	return result, nil
}

func newMetrics(metrics []v1alpha1.Metric) widget.MetricList {
	result := make([]widget.Metric, len(metrics))
	for i, m := range metrics {
		result[i] = widget.Metric{
			Name:   m.Name,
			Values: m.Values,
		}
	}

	return result
}