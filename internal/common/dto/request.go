package dto

type PagingParamsDTO struct {
	Page    int `query:"page"     validate:"min=1"`        // query param "page"
	PerPage int `query:"perPage" validate:"min=1,max=100"` // query param "per_page"
}

// ParamsToOffset converts page/perPage into SQL OFFSET and LIMIT values.
func ParamsToOffset(params PagingParamsDTO) (offset, limit int) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	return (params.Page - 1) * params.PerPage, params.PerPage
}