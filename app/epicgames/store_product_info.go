package epicgames

import (
	"encoding/json"
	"net/url"
	"time"
)

type StoreProductInfo struct {
	ProductRatings struct {
		Type string `json:"_type"`
	} `json:"productRatings"`
	DisableNewAddons      bool   `json:"disableNewAddons"`
	ModMarketplaceEnabled bool   `json:"modMarketplaceEnabled"`
	Title                 string `json:"_title"`
	RegionBlock           string `json:"regionBlock"`
	NoIndex               bool   `json:"_noIndex"`
	ProductName           string `json:"productName"`
	PageTheme             struct {
		PreferredMode string `json:"preferredMode"`
		Light         struct {
			Type string `json:"_type"`
		} `json:"light"`
		Type string `json:"_type"`
		Dark struct {
			Type  string `json:"_type"`
			Theme string `json:"theme"`
		} `json:"dark"`
	} `json:"pageTheme"`
	Namespace string `json:"namespace"`
	Theme     struct {
		Type string `json:"_type"`
	} `json:"theme"`
	ReviewOptOut     bool `json:"reviewOptOut"`
	ExternalNavLinks struct {
		Type string `json:"_type"`
	} `json:"externalNavLinks"`
	UrlPattern   string    `json:"_urlPattern"`
	Slug         string    `json:"_slug"`
	ActiveDate   time.Time `json:"_activeDate"`
	LastModified time.Time `json:"lastModified"`
	Locale       string    `json:"_locale"`
	Id           string    `json:"_id"`
	TemplateName string    `json:"_templateName"`
	Pages        []struct {
		ProductRatings struct {
			Type string `json:"_type"`
		} `json:"productRatings"`
		DisableNewAddons      bool   `json:"disableNewAddons"`
		ModMarketplaceEnabled bool   `json:"modMarketplaceEnabled"`
		Title                 string `json:"_title"`
		RegionBlock           string `json:"regionBlock"`
		NoIndex               bool   `json:"_noIndex"`
		ProductName           string `json:"productName"`
		PageTheme             struct {
			PreferredMode string `json:"preferredMode"`
			Light         struct {
				Type string `json:"_type"`
			} `json:"light"`
			Type string `json:"_type"`
			Dark struct {
				Type  string `json:"_type"`
				Theme string `json:"theme"`
			} `json:"dark"`
		} `json:"pageTheme"`
		Namespace string `json:"namespace"`
		Theme     struct {
			Type string `json:"_type"`
		} `json:"theme"`
		ReviewOptOut     bool `json:"reviewOptOut"`
		ExternalNavLinks struct {
			Type string `json:"_type"`
		} `json:"externalNavLinks"`
		UrlPattern   string                 `json:"_urlPattern"`
		Slug         string                 `json:"_slug"`
		ActiveDate   time.Time              `json:"_activeDate"`
		LastModified time.Time              `json:"lastModified"`
		Locale       string                 `json:"_locale"`
		Id           string                 `json:"_id"`
		TemplateName string                 `json:"_templateName"`
		Offer        *StoreProductInfoOffer `json:"offer"`
		Item         struct {
			Type    string `json:"_type"`
			HasItem bool   `json:"hasItem"`
		} `json:"item"`
		Data struct {
			ProductLinks struct {
				Type string `json:"_type"`
			} `json:"productLinks"`
			SocialLinks struct {
				LinkTwitter   string `json:"linkTwitter,omitempty"`
				LinkTwitch    string `json:"linkTwitch,omitempty"`
				LinkFacebook  string `json:"linkFacebook,omitempty"`
				LinkYoutube   string `json:"linkYoutube,omitempty"`
				Type          string `json:"_type"`
				LinkDiscord   string `json:"linkDiscord,omitempty"`
				LinkReddit    string `json:"linkReddit,omitempty"`
				LinkHomepage  string `json:"linkHomepage,omitempty"`
				LinkInstagram string `json:"linkInstagram,omitempty"`
			} `json:"socialLinks"`
			Requirements struct {
				Languages []string `json:"languages,omitempty"`
				Systems   []struct {
					Type       string `json:"_type"`
					SystemType string `json:"systemType"`
					Details    []struct {
						Type        string `json:"_type"`
						Title       string `json:"title"`
						Minimum     string `json:"minimum"`
						Recommended string `json:"recommended"`
					} `json:"details"`
				} `json:"systems,omitempty"`
				AccountRequirements string `json:"accountRequirements"`
				Type                string `json:"_type"`
				Rating              struct {
					Type string `json:"_type"`
				} `json:"rating"`
			} `json:"requirements"`
			NavOrder int `json:"navOrder"`
			Footer   struct {
				Type              string `json:"_type"`
				Copy              string `json:"copy,omitempty"`
				PrivacyPolicyLink struct {
					Src   string `json:"src,omitempty"`
					Type  string `json:"_type"`
					Title string `json:"title,omitempty"`
				} `json:"privacyPolicyLink"`
			} `json:"footer"`
			Type  string `json:"_type"`
			About struct {
				Image struct {
					Src  string `json:"src,omitempty"`
					Type string `json:"_type"`
				} `json:"image"`
				DeveloperAttribution string `json:"developerAttribution,omitempty"`
				Type                 string `json:"_type"`
				PublisherAttribution string `json:"publisherAttribution,omitempty"`
				Description          string `json:"description,omitempty"`
				ShortDescription     string `json:"shortDescription,omitempty"`
				Title                string `json:"title,omitempty"`
				DeveloperLogo        struct {
					Type string `json:"_type"`
				} `json:"developerLogo"`
			} `json:"about"`
			Banner struct {
				ShowPromotion bool   `json:"showPromotion"`
				Type          string `json:"_type"`
				Link          struct {
					Type string `json:"_type"`
				} `json:"link"`
			} `json:"banner"`
			Hero struct {
				LogoImage struct {
					AltText string `json:"altText,omitempty"`
					Src     string `json:"src,omitempty"`
					Type    string `json:"_type"`
				} `json:"logoImage"`
				PortraitBackgroundImageUrl string `json:"portraitBackgroundImageUrl,omitempty"`
				Type                       string `json:"_type"`
				Action                     struct {
					Type string `json:"_type"`
				} `json:"action"`
				Video struct {
					Loop          bool   `json:"loop"`
					Type          string `json:"_type"`
					HasFullScreen bool   `json:"hasFullScreen"`
					HasControls   bool   `json:"hasControls"`
					Muted         bool   `json:"muted"`
					Autoplay      bool   `json:"autoplay"`
				} `json:"video"`
				IsFullBleed        bool   `json:"isFullBleed"`
				AltContentPosition bool   `json:"altContentPosition"`
				BackgroundImageUrl string `json:"backgroundImageUrl,omitempty"`
			} `json:"hero"`
			Carousel struct {
				Type  string `json:"_type"`
				Items []struct {
					Image struct {
						Type string `json:"_type"`
					} `json:"image"`
					Type  string `json:"_type"`
					Video struct {
						Recipes       string `json:"recipes"`
						Loop          bool   `json:"loop"`
						Type          string `json:"_type"`
						HasFullScreen bool   `json:"hasFullScreen"`
						HasControls   bool   `json:"hasControls"`
						Title         string `json:"title"`
						Type1         string `json:"type"`
						Muted         bool   `json:"muted"`
						Poster        string `json:"poster"`
						Autoplay      bool   `json:"autoplay"`
					} `json:"video"`
				} `json:"items,omitempty"`
			} `json:"carousel"`
			Editions struct {
				Type         string `json:"_type"`
				EnableImages bool   `json:"enableImages"`
			} `json:"editions"`
			Meta struct {
				ReleaseDate time.Time `json:"releaseDate,omitempty"`
				Type        string    `json:"_type"`
				Publisher   []string  `json:"publisher,omitempty"`
				Logo        struct {
					Type string `json:"_type"`
				} `json:"logo"`
				Developer         []string `json:"developer,omitempty"`
				CustomReleaseDate string   `json:"customReleaseDate,omitempty"`
				Platform          []string `json:"platform,omitempty"`
				Tags              []string `json:"tags,omitempty"`
			} `json:"meta"`
			Markdown struct {
				Type    string `json:"_type"`
				Title   string `json:"title,omitempty"`
				Content string `json:"content,omitempty"`
			} `json:"markdown"`
			Dlc struct {
				ContingentOffer struct {
					RegionRestrictions struct {
						Type string `json:"_type"`
					} `json:"regionRestrictions"`
					Type     string `json:"_type"`
					HasOffer bool   `json:"hasOffer"`
				} `json:"contingentOffer"`
				Type         string `json:"_type"`
				EnableImages bool   `json:"enableImages"`
			} `json:"dlc"`
			Seo struct {
				Image struct {
					Type string `json:"_type"`
				} `json:"image"`
				Twitter struct {
					Type string `json:"_type"`
				} `json:"twitter"`
				Type string `json:"_type"`
				Og   struct {
					Type string `json:"_type"`
				} `json:"og"`
			} `json:"seo"`
			ProductSections []struct {
				ProductSection string `json:"productSection"`
				Type           string `json:"_type"`
			} `json:"productSections"`
			Gallery struct {
				Type string `json:"_type"`
			} `json:"gallery"`
			NavTitle string `json:"navTitle"`
		} `json:"data"`
		PageRegionBlock string `json:"pageRegionBlock,omitempty"`
		AgeGate         struct {
			HasAgeGate bool   `json:"hasAgeGate"`
			Type       string `json:"_type"`
		} `json:"ageGate"`
		Type   string   `json:"type"`
		Images []string `json:"_images_,omitempty"`
		Tag    string   `json:"tag,omitempty"`
	} `json:"pages"`
}

type StoreProductInfoOffer struct {
	RegionRestrictions struct {
		Type string `json:"_type"`
	} `json:"regionRestrictions"`
	Type      string `json:"_type"`
	Namespace string `json:"namespace,omitempty"`
	Id        string `json:"id,omitempty"`
	HasOffer  bool   `json:"hasOffer"`
}

const productStoreInfoUrl = "https://store-content-ipv4.ak.epicgames.com/api/en-US/content/products/"

func FetchStoreProductInfo(slug string) (*StoreProductInfo, error) {
	reqUrl, _ := url.Parse(productStoreInfoUrl)
	reqUrl = reqUrl.JoinPath(slug)

	resp, err := getClient().Get(reqUrl.String())
	if err != nil {
		return nil, err
	}

	result := new(StoreProductInfo)
	err = json.NewDecoder(resp.Body).Decode(&result)

	return result, err
}
