package dbrt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

const baseURL string = "https://www.dbrt.hu/sources/process/meroallas_ajax.php"

type fizkodResponse struct {
	MerokSzama int    `json:"merok_szama"`
	Message    string `json:"message"`
	Merok      struct {
		Mero1 mero `json:"1"`
	} `json:"mero"`
	Fizkod    string `json:"fizkod"`
	Hiba      string `json:"hiba"`
	Diktalhat string `json:"diktalhat"`
}

type mero struct {
	GyariSzam         string `json:"gyariSzam"`
	MhKod             string `json:"mhKod"`
	AktualisMeroallas string `json:"aktualisMeroallas"`
	ElozoMeroallas    string `json:"elozoMeroallas"`
	ElozoDatum        string `json:"elozoDatum"`
}

func fizkodCall(fizkod string) (fizkodResponse, error) {
	var fizkodResp fizkodResponse
	payload := strings.NewReader(fmt.Sprintf("fizkod=%s&action=fizkod", fizkod))
	req, err := http.NewRequest(http.MethodPost, baseURL, payload)
	if err != nil {
		return fizkodResp, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fizkodResp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fizkodResp, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fizkodResp, err
	}
	if string(body) == "null" {
		return fizkodResp, fmt.Errorf("invalid fizkod: %s", fizkod)
	}

	err = json.Unmarshal(body, &fizkodResp)
	return fizkodResp, err
}

type meroallasResponse struct {
	Result struct {
		Mero1 struct {
			Hiba               string `json:"hiba"`
			Figy               string `json:"figy"`
			MeroallasElfogadva string `json:"meroallasElfogadva"`
			Message            string `json:"message"`
		} `json:"1"`
	} `json:"result"`
}

func meroallasCall(formdata map[string][]byte) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range formdata {
		fw, err := w.CreateFormField(key)
		if err != nil {
			return err
		}
		_, err = fw.Write(r)
		if err != nil {
			return err
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", baseURL, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var meroallasResp meroallasResponse
	err = json.Unmarshal(body, &meroallasResp)
	if err != nil {
		return err
	}

	if meroallasResp.Result.Mero1.MeroallasElfogadva != "true" {
		return fmt.Errorf("reading wasn't accepted; warning: %s; error: %s; message: %s",
			meroallasResp.Result.Mero1.Figy,
			meroallasResp.Result.Mero1.Hiba,
			meroallasResp.Result.Mero1.Message,
		)
	}
	return nil
}

func ReportMeterReading(fizkod string, reading int) error {
	fizkodResp, err := fizkodCall(fizkod)
	if err != nil {
		return err
	}

	if fizkodResp.Diktalhat != "true" {
		message := strings.TrimPrefix(fizkodResp.Message, "<span class=\"fail\">")
		message = strings.TrimSuffix(message, "</span>")
		return fmt.Errorf(message)
	}

	formdata := map[string][]byte{
		"action":           []byte("meroallas"),
		"merok_szama":      []byte(strconv.Itoa(fizkodResp.MerokSzama)),
		"fizkod":           []byte(fizkod),
		"gyariSzam_1":      []byte(fizkodResp.Merok.Mero1.GyariSzam),
		"elozoMeroallas_1": []byte(fizkodResp.Merok.Mero1.ElozoMeroallas),
		"mhKod_1":          []byte(fizkodResp.Merok.Mero1.MhKod),
		"meroallas_1":      []byte(strconv.Itoa(207)),
	}
	err = meroallasCall(formdata)
	if err != nil {
		return err
	}

	fizkodResp, err = fizkodCall(fizkod)
	if err != nil {
		return err
	}
	if fizkodResp.Merok.Mero1.AktualisMeroallas != strconv.Itoa(reading) {
		return fmt.Errorf("reading wasn't updated")
	}
	return nil
}
