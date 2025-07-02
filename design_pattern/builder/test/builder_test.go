package test

import (
	"fmt"
	"testing"

	"github.com/msyamsula/portofolio/design_pattern/builder"
	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {

	builder := builder.NewInfraBuilder()
	infra, err := builder.SetRegion("asia").SetAutoscale(false).SetNetworks("network").SetSecurityGroups([]string{"a", "b"}).Build()
	assert.Equal(t, nil, err)
	fmt.Println(infra)

	infra, err = builder.SetRegion("").Build()
	assert.NotNil(t, err)
	fmt.Println(infra)
}
