package hotspider

// QuarkTelevisionsModel 夸克影视排行榜 API 响应模型。
type QuarkTelevisionsModel struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		Hits struct {
			Hit struct {
				WisdomExtInfo struct {
					TriggerType   string `json:"trigger_type"`
					TriggerDetail struct {
						TriggerDict     string `json:"trigger_dict"`
						TriggerEntity   string `json:"trigger_entity"`
						TriggerModuleId string `json:"trigger_module_id"`
					} `json:"trigger_detail"`
				} `json:"wisdom_ext_info"`
				Item []struct {
					Area         string `json:"area"`
					HotScore     string `json:"hot_score"`
					EpisodeCount string `json:"episode_count,omitempty"`
					Color        string `json:"color"`
					Src          string `json:"src"`
					Year         string `json:"year"`
					ContentId    string `json:"content_id"`
					Channel      string `json:"channel"`
					IsAuth       string `json:"is_auth"`
					Title        string `json:"title"`
					ScoreAvg     string `json:"score_avg"`
					LastHotScore string `json:"last_hot_score"`
					PubDate      string `json:"pub_date"`
					Actors       string `json:"actors"`
					PlayLink     string `json:"play_link"`
					Ranking      string `json:"ranking"`
					Text         string `json:"text"`
					Category     string `json:"category"`
					HotTrend     string `json:"hot_trend"`
					Desc         string `json:"desc"`
				} `json:"item"`
				ScLayout       string `json:"sc_layout"`
				WisdomHy       string `json:"wisdom_hy"`
				WisdomSc       string `json:"wisdom_sc"`
				MakeUpPriority string `json:"MakeUpPriority"`
				ScStype        string `json:"sc_stype"`
				ScExt          struct {
					Source string `json:"source"`
				} `json:"sc_ext"`
				Filter struct {
					Categorys struct {
						Category []struct {
							Name   string `json:"name"`
							Active string `json:"active,omitempty"`
						} `json:"category"`
					} `json:"categorys"`
					Ranks struct {
						Rank []struct {
							Name   string `json:"name"`
							Active string `json:"active,omitempty"`
						} `json:"rank"`
					} `json:"ranks"`
					Areas struct {
						Area []struct {
							Name   string `json:"name"`
							Active string `json:"active,omitempty"`
						} `json:"area"`
					} `json:"areas"`
					Years struct {
						Year []struct {
							Name   string `json:"name"`
							Active string `json:"active,omitempty"`
						} `json:"year"`
					} `json:"years"`
				} `json:"filter"`
				Bucket       string `json:"bucket"`
				POSITION     string `json:"POSITION"`
				WisdomModule string `json:"wisdom_module"`
				Channels     struct {
					Channel []struct {
						Name   string `json:"name"`
						Active string `json:"active,omitempty"`
					} `json:"channel"`
				} `json:"channels"`
				TotalHits  string `json:"totalHits"`
				MakeUpPos  string `json:"MakeUpPos"`
				TemplateId string `json:"template_id"`
				DiluPath   string `json:"dilu_path"`
				Status     string `json:"status"`
			} `json:"hit"`
			Numhits   string `json:"numhits"`
			Totalhits string `json:"totalhits"`
		} `json:"hits"`
	} `json:"data"`
}
