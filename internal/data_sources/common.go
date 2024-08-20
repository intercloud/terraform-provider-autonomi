package datasources

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var int64MapAttr = schema.MapAttribute{
	ElementType: types.Int64Type,
	Computed:    true,
}
