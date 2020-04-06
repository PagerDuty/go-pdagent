package eventsapi

import "gopkg.in/h2non/gock.v1"

func mockEndpointV1(statusCode int, response interface{}) *gock.Response {
	mock := gock.New("https://events.pagerduty.com").
		Post("/generic/2010-04-15/create_event.json").
		Reply(statusCode)

	if response != nil {
		mock = mock.JSON(response)
	}

	return mock
}

func mockEndpointV2(statusCode int, response interface{}) *gock.Response {
	mock := gock.New("https://events.pagerduty.com").
		Post("/v2/enqueue").
		Reply(statusCode)

	if response != nil {
		mock = mock.JSON(response)
	}

	return mock
}
