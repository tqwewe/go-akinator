package akinator

import (
	"io/ioutil"
	"net/url"
	"strconv"
)

// AnswerYes responds to a question with the answer "Yes".
func (r *Response) AnswerYes() error {
	return r.answer(0)
}

// AnswerNo responds to a question with the answer "No".
func (r *Response) AnswerNo() error {
	return r.answer(1)
}

// AnswerDontKnow responds to a question with the answer "Don't Know".
func (r *Response) AnswerDontKnow() error {
	return r.answer(2)
}

// AnswerProbably responds to a question with the answer "Probably".
func (r *Response) AnswerProbably() error {
	return r.answer(3)
}

// AnswerProbablyNot responds to a question with the answer "Probably Not".
func (r *Response) AnswerProbablyNot() error {
	return r.answer(4)
}

func (r *Response) answer(a int) error {
	resp, err := r.client.HTTPClient.Get(apiURL + "/ws/answer?" + url.Values{
		"session":   {r.client.identification.session},
		"signature": {r.client.identification.signature},
		"step":      {strconv.Itoa(r.client.identification.step)},
		"answer":    {strconv.Itoa(a)},
	}.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	newResponse, err := r.client.getResponse(body)
	if err != nil {
		return err
	}

	r.client.responses <- newResponse

	return nil
}
