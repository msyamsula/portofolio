package builder

import "errors"

type InfraBuilder struct {
	Infra *Infrastructure
}

type Infrastructure struct {
	Region         string
	InstanceType   string
	StorageOption  string
	Networks       string
	SecurityGroups []string
	Tags           []string
	Autoscale      bool
}

func NewInfraBuilder() *InfraBuilder {
	return &InfraBuilder{
		Infra: new(Infrastructure),
	}
}
func (i *InfraBuilder) SetRegion(r string) *InfraBuilder {
	i.Infra.Region = r
	return i
}
func (i *InfraBuilder) SetInstanceType(t string) *InfraBuilder {
	i.Infra.InstanceType = t
	return i
}
func (i *InfraBuilder) SetStorage(s string) *InfraBuilder {
	i.Infra.StorageOption = s
	return i
}
func (i *InfraBuilder) SetNetworks(s string) *InfraBuilder {
	i.Infra.Networks = s
	return i
}
func (i *InfraBuilder) SetSecurityGroups(groups []string) *InfraBuilder {
	i.Infra.SecurityGroups = groups
	return i
}
func (i *InfraBuilder) AddSecurityGroup(group string) *InfraBuilder {
	i.Infra.SecurityGroups = append(i.Infra.SecurityGroups, group)
	return i
}
func (i *InfraBuilder) SetTags(tags []string) *InfraBuilder {
	i.Infra.Tags = tags
	return i
}
func (i *InfraBuilder) SetAutoscale(s bool) *InfraBuilder {
	i.Infra.Autoscale = s
	return i
}
func (i *InfraBuilder) Build() (*Infrastructure, error) {
	if i.Infra.Region == "" || len(i.Infra.SecurityGroups) == 0 {
		return nil, errors.New("region and security groups are not set")
	}

	return i.Infra, nil
}
