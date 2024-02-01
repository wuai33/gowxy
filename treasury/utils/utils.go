package utils


//Send http request to lcmaas and check response code
func SendRequestAndCheckResponseCode(client HttpClient, method, url string, data io.Reader, expectReturnCode int) ([]byte, bool, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, false, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode == expectReturnCode {
		return out, true, nil
	} else {
		return out, false, nil
	}
}