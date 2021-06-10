/*
 * Copyright 2021 Red Hat
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package chartverifier

import (
	"github.com/redhat-certification/chart-verifier/pkg/chartverifier/checks"
	"github.com/spf13/viper"
	helmcli "helm.sh/helm/v3/pkg/cli"
)

type CheckNotFoundErr string

func (e CheckNotFoundErr) Error() string {
	return "check not found: " + string(e)
}

type CheckErr string

func (e CheckErr) Error() string {
	return "check error: " + string(e)
}

func NewCheckErr(err error) error {
	return CheckErr(err.Error())
}

type AnnotationHolder struct {
	Holder                        ReportBuilder
	CertifiedOpenShiftVersionFlag string
}

func (holder *AnnotationHolder) SetCertifiedOpenShiftVersion(version string) {
	holder.Holder.SetCertifiedOpenShiftVersion(version)
}

func (holder *AnnotationHolder) GetCertifiedOpenShiftVersionFlag() string {
	return holder.CertifiedOpenShiftVersionFlag
}

type verifier struct {
	config           *viper.Viper
	registry         checks.Registry
	requiredChecks   []string
	settings         *helmcli.EnvSettings
	toolVersion      string
	openshiftVersion string
	values           map[string]interface{}
}

func (c *verifier) subConfig(name string) *viper.Viper {
	if sub := c.config.Sub(name); sub == nil {
		return viper.New()
	} else {
		return sub
	}
}

func (c *verifier) Verify(uri string) (*Report, error) {

	chrt, _, err := checks.LoadChartFromURI(uri)
	if err != nil {
		return nil, err
	}

	result := NewReportBuilder().
		SetToolVersion(c.toolVersion).
		SetChartUri(uri).
		SetChart(chrt)

	for _, name := range c.requiredChecks {
		check, ok := c.registry.Get(name)
		if !ok {
			return nil, CheckNotFoundErr(name)
		}

		holder := AnnotationHolder{Holder: result,
			CertifiedOpenShiftVersionFlag: c.openshiftVersion}

		r, checkErr := check.Func(&checks.CheckOptions{
			HelmEnvSettings:  c.settings,
			URI:              uri,
			Values:           c.values,
			ViperConfig:      c.subConfig(name),
			AnnotationHolder: &holder,
		})

		if checkErr != nil {
			return nil, NewCheckErr(checkErr)
		}
		_ = result.AddCheck(name, check.Type, r)

	}

	return result.Build()
}
