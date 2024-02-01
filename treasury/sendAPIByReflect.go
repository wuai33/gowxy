package treasury


var (
	ep = "localhot:8080"
	baseURL = "/v1"
)

type Test struct {
	Name    string `json:"name"`
}

// A general template function for API Path generation and sending Request.
// Define the path corresponding to the API according to the model of the object.
// For example: for Test, the api is http://localhot:8080/v1/tests
func UniversalRequestSend(client HttpClient, method, url, param string, expectReturnCode int, obj *interface{}) (*models.ErrorBody, error) {
	var reqData io.Reader
	t := reflect.TypeOf(*obj).Elem()
	objName := t.Name()

	lowerPlural := strings.ToLower(objName) + "s"
	apiPath := url + lowerPlural
	if param != "" {
		apiPath = apiPath + "/" + param
	}
	if method == http.MethodGet || method == http.MethodDelete {
		reqData = nil
	} else {
		data, err := json.Marshal(*obj)
		if err != nil {
			return nil, err
		}
		reqData = bytes.NewBuffer(data)
	}

	out, expect, err := SendRequestAndCheckResponseCode(client, method, apiPath, reqData, expectReturnCode)
	if err != nil {
		return nil, err
	} else {
		if expect {
			if method == http.MethodDelete {
				return nil, nil
			}
			outVal := reflect.ValueOf(out)
			objVal := reflect.ValueOf(obj)
			reflect.Indirect(objVal).Set(reflect.Indirect(outVal))
			return nil, nil
		} else {
			errorinfo := models.ErrorBody{}
			err = json.Unmarshal(out, &errorinfo)
			if err != nil {
				return nil, err
			}
			*obj = nil
			return &errorinfo, nil
		}
	}
}



func TestCreateRequest(test *Test) (*Test, *models.ErrorBody, error) {
	serverURL := "http://" + ep + baseURL
	testObj := interface{}(test)
	errBody, err := util.UniversalRequestSend(Client, http.MethodPost, serverURL, "", http.StatusCreated, &testObj)
	if testObj != nil {
		testByte := testObj.([]byte)
		_ = json.Unmarshal(testByte, test)
		return test, errBody, err
	}
	return nil, errBody, err
}


func TestPatchRequest(test *Test) (*Test, *models.ErrorBody, error) {
	serverURL := "http://" + ep + baseURL
	testObj := interface{}(test)
	errBody, err := util.UniversalRequestSend(Client, http.MethodPatch, serverURL, testName, http.StatusOK, &testObj)
	if testObj != nil {
		testByte := testObj.([]byte)
		_ = json.Unmarshal(testByte, test)
		return test, errBody, err
	}
	return nil, errBody, err
}


func TestDeleteRequest(test *Test) (*Test, *models.ErrorBody, error) {
	serverURL := "http://" + ep + baseURL
	testObj := interface{}(&Test{})
	errBody, err := util.UniversalRequestSend(Client, http.MethodDelete, serverURL, name, http.StatusNoContent, &testObj)
	return errBody, err
	
}

func TestGetRequest(test *Test) (*Test, *models.ErrorBody, error) {
	serverURL := "http://" + ep + baseURL
	var test Test
	testObj := interface{}(&test)
	errBody, err := util.UniversalRequestSend(Client, http.MethodGet, serverURL, name, http.StatusOK, &testObj)
	if testObj != nil {
		testByte := testObj.([]byte)
		_ = json.Unmarshal(testByte, &test)
		return &test, errBody, err
	}
	return nil, errBody, err
}

func TestListRequest(test *Test) (*Test, *models.ErrorBody, error) {
	serverURL := "http://" + ep + baseURL
	var tests []Test
	testObj := interface{}(tests)
	errBody, err := util.UniversalRequestSend(Client, http.MethodGet, serverURL, "", http.StatusOK, &testObj)
	if testObj != nil {
		testByte := testObj.([]byte)
		_ = json.Unmarshal(testByte, &tests)
		return tests, errBody, err
	}
	return nil, errBody, err
	
}