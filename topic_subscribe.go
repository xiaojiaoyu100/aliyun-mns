package alimns

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func checkTopicName(name string) bool {
	regex := regexp.MustCompile(`^[0-9a-zA-Z]{1}[0-9a-zA-Z-]{0,255}$`)
	return regex.MatchString(name)
}

func checkSubscription(name string) bool {
	regex := regexp.MustCompile(`^[0-9a-zA-Z]{1}[0-9a-zA-Z-]{0,255}$`)
	return regex.MatchString(name)
}

// Subscribe 订阅主题
func (c *Client) Subscribe(topic, subscriptions string, endpoint EndPoint, setters ...SubscribeParamSetter) error {
	var err error

	if !checkTopicName(topic) {
		return errors.New("unqualified topic name")
	}

	if !checkSubscription(subscriptions) {
		return errors.New("unqualified Subscriptio nname")
	}

	ep, err := endpoint.EndPoint()
	if err != nil {
		return errors.New("invalid endpoint")
	}

	subscribeParam := defaultSubscribeParam

	for _, setter := range setters {
		err = setter(&subscribeParam)
		if err != nil {
			return err
		}
	}

	subscribeParam.EndPoint = ep
	requestLine := fmt.Sprintf(mnsSubscribe, topic, subscriptions)

	req := c.ca.NewRequest().Put().WithPath(requestLine).WithXMLBody(&subscribeParam).WithTimeout(apiTimeout)
	resp, err := c.ca.Do(req)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusCreated, http.StatusNoContent:
		return nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil
		}

		switch respErr.Code {
		case subscriptionNameLengthError.Error():
			return subscriptionNameLengthError
		case subscriptionNameInvalidError.Error():
			return subscriptionNameInvalidError
		case subscriptionAlreadyExistError.Error():
			return subscriptionAlreadyExistError
		case endpointInvalidError.Error():
			return endpointInvalidError
		case invalidArgumentError.Error():
			return invalidArgumentError
		default:
			return errors.New(respErr.Code)
		}
	}
}
