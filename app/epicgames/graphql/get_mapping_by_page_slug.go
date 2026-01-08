package graphql

type Mapping struct {
	SandboxId string `json:"sandboxId"`
	ProductId string `json:"productId"`
	Mappings  struct {
		CmsSlug interface{} `json:"cmsSlug"`
		OfferId string      `json:"offerId"`
		Offer   struct {
			Id        string `json:"id"`
			Namespace string `json:"namespace"`
		} `json:"offer"`
	} `json:"mappings"`
}

type getMappingByPageSlugResponse struct {
	Data struct {
		StorePageMapping struct {
			Mapping *Mapping `json:"mapping"`
		} `json:"StorePageMapping"`
	} `json:"data"`
}

const getMappingsByPageSlugExtensions extensions = "{\"persistedQuery\":{\"version\":1,\"sha256Hash\":\"781fd69ec8116125fa8dc245c0838198cdf5283e31647d08dfa27f45ee8b1f30\"}}"

func GetMappingByPageSlug(locale, slug string) (*Mapping, error) {
	response := getMappingByPageSlugResponse{}

	err := doRequest(
		graphqlRequest{
			operationName: "getMappingByPageSlug",
			variables: map[string]any{
				"pageSlug": slug,
				"locale":   locale,
			},
			extensions: getMappingsByPageSlugExtensions,
		},
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.Data.StorePageMapping.Mapping, nil
}
