package models

// RequestSentLinks - сущность для приема ссылок на проверку.
type RequestSentLinks struct {
	Links []string `json:"links"`
}

// ResponseSentLinks - структура для выдачи обработанных ссылок.
type ResponseSentLinks struct {
	Links map[string]string `json:"links"`
	Num   int               `json:"links_num"`
}

// RequestLinksNum - сущность для получения запроса на выдачу ссылок.
type RequestLinksNum struct {
	LinksList []int `json:"links_list"`
}
