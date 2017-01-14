package akinator

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strconv"
)

// Response contains all the information responded by the Akinator.
// If a Response is a guess, the Guessed key will be set to true,
// otherwise if it is a question, it will be set to false.
type Response struct {
	Guessed              bool
	CharacterName        string
	CharacterDescription string
	CharacterImageURL    string
	CharacterProbability float64

	Akitude        string
	Question       string
	Emotion        int
	Status         string
	Progression    float64
	client         *Client
	stepOfLastProp int
}

func (c *Client) getResponse(content []byte) (*Response, error) {
	var (
		r   Response
		err error
	)

	r.client = c

	var decoded struct {
		Completion string `json:"completion"`
		Parameters struct {
			Identification struct {
				Channel   int    `json:"channel"`
				Session   string `json:"session"`
				Signature string `json:"signature"`
			} `json:"identification"`
			StepInformation struct {
				Answers []struct {
					Answer string `json:"answer"`
				} `json:"answers"`
				Infogain    string `json:"infogain"`
				Progression string `json:"progression"`
				Question    string `json:"question"`
				Questionid  string `json:"questionid"`
				Step        string `json:"step"`
			} `json:"step_information"`
			Answers []struct {
				Answer string `json:"answer"`
			} `json:"answers"`
			Infogain       int    `json:"infogain"`
			Progression    string `json:"progression"`
			Question       string `json:"question"`
			Questionid     string `json:"questionid"`
			StatusMinibase string `json:"status_minibase"`
			Step           string `json:"step"`
		} `json:"parameters"`
	}

	if err = json.Unmarshal(content, &decoded); err != nil {
		return &r, err
	}

	r.Status = decoded.Completion

	if decoded.Parameters.Identification.Channel != 0 || decoded.Parameters.Identification.Session != "" || decoded.Parameters.Identification.Signature != "" {
		if step, err := strconv.Atoi(decoded.Parameters.StepInformation.Step); err != nil {
			c.identification.step++
		} else {
			c.identification.step = step
		}
		c.identification.session = decoded.Parameters.Identification.Session
		c.identification.signature = decoded.Parameters.Identification.Signature
		if progress, err := strconv.ParseFloat(decoded.Parameters.StepInformation.Progression, 64); err == nil {
			r.Progression = progress
		}
		r.Question = decoded.Parameters.StepInformation.Question
	} else {
		if step, err := strconv.Atoi(decoded.Parameters.Step); err != nil {
			c.identification.step++
		} else {
			c.identification.step = step
		}
		if progress, err := strconv.ParseFloat(decoded.Parameters.Progression, 64); err == nil {
			r.Progression = progress
		}
		r.Question = decoded.Parameters.Question
	}

	r.Akitude = r.getAkitude()
	c.previousProgress = r.Progression

	if r.isAbleToFind() {
		r.stepOfLastProp = c.identification.step

		resp, err := c.HTTPClient.Get("http://api-us4.akinator.com/ws/list?" + url.Values{
			"session":        {c.identification.session},
			"signature":      {c.identification.signature},
			"step":           {strconv.Itoa(c.identification.step)},
			"size":           {"2"},
			"mode_question":  {"0"},
			"max_pic_width":  {"246"},
			"max_pic_height": {"294"},
			"pref_photos":    {"VO-OK"},
		}.Encode())
		if err != nil {
			return &r, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &r, err
		}

		var guess struct {
			Completion string `json:"completion"`
			Parameters struct {
				NbObjetsPertinents string `json:"NbObjetsPertinents"`
				Elements           []struct {
					Element struct {
						AbsolutePicturePath string `json:"absolute_picture_path"`
						Description         string `json:"description"`
						ID                  string `json:"id"`
						IDBase              string `json:"id_base"`
						MinibaseAddable     string `json:"minibase_addable"`
						Name                string `json:"name"`
						PicturePath         string `json:"picture_path"`
						Proba               string `json:"proba"`
						Pseudo              string `json:"pseudo"`
						Ranking             string `json:"ranking"`
						RelativeID          string `json:"relative_id"`
						ValideContrainte    string `json:"valide_contrainte"`
					} `json:"element"`
				} `json:"elements"`
			} `json:"parameters"`
		}

		if err = json.Unmarshal(body, &guess); err != nil {
			return &r, err
		}

		if len(guess.Parameters.Elements) >= 1 {
			r.Guessed = true
			r.CharacterName = guess.Parameters.Elements[0].Element.Name
			r.CharacterDescription = guess.Parameters.Elements[0].Element.Description
			r.CharacterImageURL = guess.Parameters.Elements[0].Element.AbsolutePicturePath
			r.CharacterProbability, _ = strconv.ParseFloat(guess.Parameters.Elements[0].Element.Proba, 64)
		}
	}

	return &r, err
}

func (r *Response) isAbleToFind() bool {
	step := r.client.identification.step

	if r.client.identification.step == 79 {
		return true
	}

	if r.Progression > 96 || step-r.stepOfLastProp == 25 {
		if r.client.identification.step != 75 {
			return true
		}
	}

	return false
}

func (r *Response) getAkitude() string {
	step := r.client.identification.step
	progress := r.Progression
	oldProgress := r.client.previousProgress
	f := float64(step) * 4
	var p float64
	if step <= 10 {
		p = (float64(step)*progress + (10-float64(step))*f) / 10
	}
	if progress >= 80 {
		return "༼ つ ◕_◕ ༽つ" // akinator_mobile.png
	}
	if oldProgress < 50 && progress >= 50 {
		return "(°ロ°)☝" // akinator_inspiration_forte.png
	}
	if progress >= 50 {
		return "✌(-‿-)✌" // akinator_confiant.png
	}
	if oldProgress-progress > 16 {
		return "◉_◉" // akinator_surprise.png
	}
	if oldProgress-progress > 8 {
		return "⚆ _ ⚆" // akinator_etonnement.png
	}
	if p >= f {
		return "(σヘσ)" // akinator_inspiration_legere.png
	}
	if float64(p) >= float64(f)*0.8 {
		return "|⸟◞ ⸟|" // akinator_serein.png
	}
	if float64(p) >= float64(f)*0.6 {
		return "ಠ_ಠ" // akinator_concentration_intense.png
	}
	if float64(p) >= float64(f)*0.4 {
		return "(ಥ_ಥ)" // akinator_leger_decouragement.png
	}
	if float64(p) >= float64(f)*0.2 {
		return "(ಥ﹏ಥ)" // akinator_tension.png
	}
	return "(ง'̀-'́)ง" // akinator_vrai_decouragement.png
}
