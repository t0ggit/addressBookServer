package addressBookService

import (
	"addressBookServer/gates/psg"
	"addressBookServer/models/dto"
	"addressBookServer/pkg"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type AddressBookService struct {
	server http.Server
	db     *psg.Psg
}

func NewAddressBookService(addr string, p *psg.Psg) (abs *AddressBookService) {
	abs = new(AddressBookService)
	abs.server = http.Server{}
	router := http.NewServeMux()
	router.HandleFunc("/create", abs.createRecordHandler)
	router.HandleFunc("/get", abs.getRecordsHandler)
	router.HandleFunc("/update", abs.updateRecordHandler)
	router.HandleFunc("/delete", abs.deleteRecordByPhoneHandler)
	abs.server.Handler = router
	abs.server.Addr = addr
	abs.db = p
	return abs
}

func (abs *AddressBookService) Start() {
	wErr := pkg.NewWrappedError("(abs *AddressBookService) Start()")

	err := abs.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		wErr.Specify(err, "abs.server.ListenAndServe()").LogError()
		return
	}
	wErr.LogMsg("Server closed")
	return
}

func (abs *AddressBookService) Close() error {
	return abs.server.Close()
}

// createRecordHandler обрабатывает запрос на создание записи
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида (все поля обязательны для заполнения):
  {"name": "Имя", "last_name": "Фамилия", "middle_name": "Отчество", "address": "Адрес", "phone": "Телефон"}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
  {"result": "OK", "data": null, "error": ""}

В случае ошибки:
  {"result": "ERROR", "data": null, "error": "error description"}
*/
func (abs *AddressBookService) createRecordHandler(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(abs *AddressBookService) createRecordHandler()")
	if err != nil {
		log.Println("(abs *AddressBookService) createRecordHandler: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Парсинг запроса
	record := dto.Record{}
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "json.Unmarshal(byteReq, &record)").LogError()
		return
	}

	// Проверка наличия необходимых данных в запросе
	if record.Name == "" || record.LastName == "" || record.Address == "" || record.Phone == "" {
		err = errors.New("required data is missing")
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(fmt.Sprintf("%s: {name: '%s', lastName: '%s', address: '%s', phone: '%s'}",
			err.Error(), record.Name, record.LastName, record.Address, record.Phone))
		return
	}

	// Нормализация номера телефона
	record.Phone, err = pkg.NormalizePhoneNumber(record.Phone)
	if err != nil {
		err = errors.New("wrong Phone")
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "pkg.NormalizePhoneNumber(record.Phone)").LogError()
		return
	}

	// Сохранение записи
	err = abs.db.SaveRecord(record)
	if err != nil {
		err = errors.New("cannot save record")
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "abs.db.SaveRecord(record)").LogError()
		return
	}

	resp.Update("OK", nil, "")
}

// updateRecordHandler обрабатывает запрос на обновление записи
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида (обязательно нужен номер и данные для обновления, т.е. номер нельзя изменить):
  {"name": "Имя", "last_name": "Фамилия", "middle_name": "Отчество", "address": "Адрес", "phone": "Телефон"}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
  {"result": "OK", "data": null, "error": ""}

В случае ошибки:
  {"result": "ERROR", "data": null, "error": "error description"}
*/
func (abs *AddressBookService) updateRecordHandler(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(abs *AddressBookService) updateRecordHandler()")
	if err != nil {
		log.Println("(abs *AddressBookService) updateRecordHandler: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Парсинг запроса
	record := dto.Record{}
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "json.Unmarshal(byteReq, &record)").LogError()
		return
	}

	// Проверка наличия необходимых данных в запросе
	if (record.Name == "" && record.LastName == "" && record.MiddleName == "" && record.Address == "") || record.Phone == "" {
		err = errors.New("required data is missing")
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(fmt.Sprintf("%s: {name: '%s', lastName: '%s', middleName: '%s', address: '%s', phone: '%s'}",
			err.Error(), record.Name, record.LastName, record.MiddleName, record.Address, record.Phone))
		return
	}

	// Нормализация номера телефона
	record.Phone, err = pkg.NormalizePhoneNumber(record.Phone)
	if err != nil {
		err = errors.New("wrong Phone")
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "pkg.NormalizePhoneNumber(record.Phone)").LogError()
		return
	}

	// Обновление записи
	err = abs.db.UpdateRecord(record)
	if err != nil {
		err = errors.New("cannot update record")
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "abs.db.UpdateRecord(record)").LogError()
		return
	}

	resp.Update("OK", nil, "")
}

// deleteRecordByPhoneHandler обрабатывает запрос на удаление записи по номеру телефона
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида:
  {"phone": "89995554422"}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
  {"result": "OK", "data": null, "error": ""}

В случае ошибки:
  {"result": "ERROR", "data": null, "error": "error description"}
*/
func (abs *AddressBookService) deleteRecordByPhoneHandler(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(abs *AddressBookService) deleteRecordByPhoneHandler()")
	if err != nil {
		log.Println("(abs *AddressBookService) deleteRecordByPhoneHandler: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Парсинг запроса
	record := dto.Record{}
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "json.Unmarshal(byteReq, &record)").LogError()
		return
	}

	// Проверка наличия необходимых данных в запросе
	if record.Phone == "" {
		err = errors.New("phone data is missing")
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(fmt.Sprintf("%s: {phone: '%s'}", err.Error(), record.Phone))
		return
	}

	// Нормализация номера телефона
	record.Phone, err = pkg.NormalizePhoneNumber(record.Phone)
	if err != nil {
		err = errors.New("wrong Phone")
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "pkg.NormalizePhoneNumber(record.Phone)").LogError()
		return
	}

	// Удаление записи
	err = abs.db.DeleteRecordByPhone(record.Phone)
	if err != nil {
		err = errors.New("cannot delete record")
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "abs.db.DeleteRecordByPhone(record.Phone)").LogError()
		return
	}

	resp.Update("OK", nil, "")
}

// getRecordsHandler обрабатывает запрос на получение записей
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида:
  {"phone": "Телефон", "name": "Имя", "last_name": "Фамилия", "middle_name": "Отчество", "address": "Адрес"}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
  {"result": "OK", "data": [ <массив записей> ], "error": ""}

В случае ошибки:
{"result": "ERROR", "data": null, "error": "error description"}
*/
func (abs *AddressBookService) getRecordsHandler(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(abs *AddressBookService) getRecordsHandler()")
	if err != nil {
		log.Println("(abs *AddressBookService) getRecordsHandler: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Парсинг запроса
	record := dto.Record{}
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "json.Marshal(req)").LogError()
		return
	}

	// Нормализация номера телефона, если указан
	if record.Phone != "" {
		record.Phone, err = pkg.NormalizePhoneNumber(record.Phone)
		if err != nil {
			resp.Update("ERROR", nil, err.Error())
			wErr.Specify(err, "pkg.NormalizePhoneNumber(record.Phone)").LogError()
			return
		}
	}

	// Получение записей
	records, err := abs.db.GetRecords(record)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "abs.db.GetRecords(record)").LogError()
		return
	}

	// Преобразование записей в формат JSON
	recordsJSON, err := json.Marshal(records)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "json.Marshal(records)").LogError()
		return
	}

	resp.Update("OK", recordsJSON, "")
}

func setHttpHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
}

func writeResponseContent(w http.ResponseWriter, resp *dto.Response, wErr *pkg.WrappedError) {
	defer wErr.Close()

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		wErr.Specify(err, "json.NewEncoder(w).Encode(resp)").LogError()
		resp.Update("ERROR", nil, "internal server error")
		return
	}
}
