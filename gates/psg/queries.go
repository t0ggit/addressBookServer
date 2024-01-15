package psg

import (
	"addressBookServer/models/dto"
	"addressBookServer/pkg"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"html/template"
	"log"
	"reflect"
	"strconv"
	"strings"
)

/*
CREATE TABLE address_book (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    last_name VARCHAR(255),
    middle_name VARCHAR(255),
    address VARCHAR(255),
    phone VARCHAR(20)
);
*/

// SaveRecord сохраняет запись в таблицу address_book. Перед сохранением
// проверяет уникальность номера телефона. Если номер телефона уже существует
// в базе данных, возвращает ошибку "phone number already in use". В случае
// успешного сохранения возвращает nil.
//
// Пример использования:
//
//	rec := dto.Record{
//	    Name:       "John",
//	    LastName:   "Doe",
//	    MiddleName: "Smith",
//	    Address:    "123 Main St",
//	    Phone:      "+71234567890",
//	}
//	err := psg.SaveRecord(rec)
//	if err != nil {
//	    fmt.Println(err.Error())
//	}
func (p *Psg) SaveRecord(rec dto.Record) error {
	wErr, err := pkg.NewWrappedErrorWithFile("(p *Psg) SaveRecord()")
	if err != nil {
		log.Println("(p *Psg) SaveRecord(): NewWrappedErrorWithFile()", err)
	}

	err = p.PhoneExists(rec.Phone)
	if err != nil {
		wErr.Specify(err, "p.PhoneExists(rec.Phone)").LogError()
		return err
	}

	sqlCommand := `INSERT INTO address_book (name, last_name, middle_name, address, phone) VALUES ($1, $2, $3, $4, $5)`
	_, err = p.conn.Exec(context.Background(), sqlCommand, rec.Name, rec.LastName, rec.MiddleName, rec.Address, rec.Phone)
	if err != nil {
		wErr.Specify(err, "p.conn.Exec()").LogError()
		return err
	}

	return nil
}

// GetRecords возвращает список записей из таблицы address_book, удовлетворяющих
// условиям выборки, определенным в переданной структуре rec. В случае успешного
// выполнения запроса возвращает список записей и nil ошибки. В случае возникновения
// ошибки при выполнении запроса или сканирования результатов, возвращает пустой
// список и ошибку.
//
// Пример использования:
//
//	rec := dto.Record{Name: "John", Phone: "+71234567890"}
//	records, err := psg.GetRecords(rec)
//	if err != nil {
//	    fmt.Println(err.Error())
//	}
func (p *Psg) GetRecords(rec dto.Record) (result []dto.Record, err error) {
	wErr, err := pkg.NewWrappedErrorWithFile("(p *Psg) GetRecords()")
	if err != nil {
		log.Println("(p *Psg) SaveRecord(): NewWrappedErrorWithFile()", err)
	}

	sqlCommand, values, err := p.SelectRecord(rec)
	if err != nil {
		wErr.Specify(err, "p.SelectRecord(rec)").LogError()
		return result, err
	}

	rows, err := p.conn.Query(context.Background(), sqlCommand, values...)
	if err != nil {
		wErr.Specify(err, "p.conn.Query()").LogError()
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var r dto.Record
		err = rows.Scan(&r.ID, &r.Name, &r.LastName, &r.MiddleName, &r.Address, &r.Phone)
		if err != nil {
			wErr.Specify(err, "rows.Scan(&r.ID, &r.Name, &r.LastName, &r.MiddleName, &r.Address, &r.Phone)").LogError()
			return result, err
		}
		result = append(result, r)
	}

	err = rows.Err()
	if err != nil {
		wErr.Specify(err, "rows.Err()").LogError()
		return result, err
	}

	return result, nil
}

// UpdateRecord обновляет запись в таблице address_book на основе переданной структуры rec.
// В случае успешного выполнения обновления возвращает nil ошибки. Если номер телефона не найден,
// возвращает ошибку "phone number not found". В случае возникновения ошибки при выполнении запроса,
// возвращает соответствующую ошибку.
//
// Пример использования:
//
//	rec := dto.Record{Phone: "+71234567890", Name: "John", Address: "123 Main St."}
//	err := psg.UpdateRecord(rec)
//	if err != nil {
//	    fmt.Println(err.Error())
//	}
func (p *Psg) UpdateRecord(rec dto.Record) (err error) {
	wErr, err := pkg.NewWrappedErrorWithFile("(p *Psg) UpdateRecord()")
	if err != nil {
		log.Println("(p *Psg) UpdateRecord(): NewWrappedErrorWithFile()", err)
	}

	err = p.PhoneExists(rec.Phone)
	if err == nil {
		err = errors.New("phone number not found")
		wErr.LogMsg(err.Error())
		return err
	}

	fields := []string{}
	values := []any{}
	index := 1

	if rec.Name != "" {
		fields = append(fields, fmt.Sprintf("name=$%d", index))
		values = append(values, rec.Name)
		index++
	}
	if rec.LastName != "" {
		fields = append(fields, fmt.Sprintf("last_name=$%d", index))
		values = append(values, rec.LastName)
		index++
	}
	if rec.MiddleName != "" {
		fields = append(fields, fmt.Sprintf("middle_name=$%d", index))
		values = append(values, rec.MiddleName)
		index++
	}
	if rec.Address != "" {
		fields = append(fields, fmt.Sprintf("address=$%d", index))
		values = append(values, rec.Address)
		index++
	}

	values = append(values, rec.Phone)

	sqlCommand := fmt.Sprintf(`UPDATE address_book SET %s WHERE phone=$%d`, strings.Join(fields, ", "), index)
	_, err = p.conn.Exec(context.Background(), sqlCommand, values...)
	if err != nil {
		wErr.Specify(err, "p.conn.Exec()").LogError()
		return err
	}

	return nil
}

// DeleteRecordByPhone удаляет запись из таблицы address_book по номеру телефона.
// В случае успешного выполнения удаления возвращает nil ошибки. Если номер телефона не найден,
// возвращает ошибку "phone number not found". В случае возникновения ошибки при выполнении запроса,
// возвращает соответствующую ошибку.
//
// Пример использования:
//
//	err := psg.DeleteRecordByPhone("+71234567890")
//	if err != nil {
//	    fmt.Println(err.Error())
//	}
func (p *Psg) DeleteRecordByPhone(phone string) (err error) {
	wErr, err := pkg.NewWrappedErrorWithFile("(p *Psg) DeleteRecordByPhone()")
	if err != nil {
		log.Println("(p *Psg) DeleteRecordByPhone(): NewWrappedErrorWithFile()", err)
	}

	err = p.PhoneExists(phone)
	if err == nil {
		err = errors.New("phone number not found")
		wErr.LogMsg(err.Error())
		return err
	}

	sqlCommand := `DELETE FROM address_book WHERE phone=$1`
	_, err = p.conn.Exec(context.Background(), sqlCommand, phone)
	if err != nil {
		wErr.Specify(err, "p.conn.Exec()").LogError()
		return err
	}

	return nil
}

// SelectRecord выполняет SQL-запрос для выборки записей из таблицы address_book
// на основе переданной структуры r, содержащей поля для условий выборки.
// Возвращает сгенерированный SQL-запрос, значения для передачи в запрос (values)
// и ошибку в случае возникновения проблем.
//
// Условия выборки строятся на основе полей структуры r, и каждое поле
// используется в качестве отдельного условия. Условия объединяются
// операторами AND, а значения подставляются через параметры $1, $2, и т.д.
//
// Пример использования:
//
//	r := dto.Record{ID: 1, Name: "John"}
//	query, values, err := psg.SelectRecord(r)
//
// Полученный query:
//
//	SELECT
//	    id, name, last_name, middle_name, address, phone
//	FROM
//	    address_book
//	WHERE
//	    id = $1 AND name = $2;
//
// Полученные значения values:
//
//	[]any{1, "John"}
func (p *Psg) SelectRecord(r dto.Record) (resQuery string, values []any, err error) {
	// Случай отсутствия условий выборки (возвращаем все записи)
	if r.Name == "" && r.LastName == "" && r.MiddleName == "" && r.Address == "" && r.Phone == "" {
		resQuery = "SELECT id, name, last_name, middle_name, address, phone FROM address_book;"
		return resQuery, values, nil
	}

	sqlFields, values, err := structToFieldsValues(r, "sql.field")
	if err != nil {
		return "", nil, err
	}

	var conds []dto.Cond

	for i := range sqlFields {
		if i == 0 {
			conds = append(conds, dto.Cond{
				Lop:    "",
				PgxInd: "$" + strconv.Itoa(i+1),
				Field:  sqlFields[i],
				Value:  values[i],
			})
			continue
		}
		conds = append(conds, dto.Cond{
			Lop:    "AND",
			PgxInd: "$" + strconv.Itoa(i+1),
			Field:  sqlFields[i],
			Value:  values[i],
		})
	}

	query := `
	SELECT 
		id, name, last_name, middle_name, address, phone
	FROM
	    address_book
	WHERE
		{{range .}} {{.Lop}} {{.Field}} = {{.PgxInd}}{{end}}
;
`
	tmpl, err := template.New("").Parse(query)
	if err != nil {
		return
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, conds)
	if err != nil {
		return
	}
	resQuery = sb.String()
	return
}

// structToFieldsValues преобразует структуру s в список имен полей и их значений,
// учитывая тег tag для определения имени поля в SQL-запросе.
//
// Возвращает два слайса: sqlFields содержит имена полей для использования в SQL-запросе,
// values содержит значения соответствующих полей. Если поле имеет значение по умолчанию
// для своего типа (например, 0 для числовых типов, "" для строк и т.д.), оно будет
// пропущено и не включено в результирующие срезы.
//
// Пример использования:
//
//	type MyStruct struct {
//	    Field1 int    `sql.field:"column1"`
//	    Field2 string `sql.field:"column2"`
//	    Field3 bool   `sql.field:"-"`
//	}
//	s := MyStruct{Field1: 42, Field2: "value", Field3: true}
//	sqlFields, values, err := structToFieldsValues(s, "sql.field")
//
// Полученные sqlFields:
//
//	[]string{"column1", "column2"}
//
// Полученные values:
//
//	[]any{42, "value"}
//
// В случае, если s не является структурой, возвращается ошибка "s must be a struct".
func structToFieldsValues(s any, tag string) (sqlFields []string, values []any, err error) {
	rv := reflect.ValueOf(s)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, nil, errors.New("s must be a struct")
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Type().Field(i)
		tg := strings.TrimSpace(field.Tag.Get(tag))
		if tg == "" || tg == "-" {
			continue
		}
		tgs := strings.Split(tg, ",")
		tg = tgs[0]

		fv := rv.Field(i)
		isZero := false
		switch fv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			isZero = fv.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			isZero = fv.Uint() == 0
		case reflect.Float32, reflect.Float64:
			isZero = fv.Float() == 0
		case reflect.Complex64, reflect.Complex128:
			isZero = fv.Complex() == complex(0, 0)
		case reflect.Bool:
			isZero = !fv.Bool()
		case reflect.String:
			isZero = fv.String() == ""
		case reflect.Array, reflect.Slice:
			isZero = fv.Len() == 0
		}

		if isZero {
			continue
		}

		sqlFields = append(sqlFields, tg)
		values = append(values, fv.Interface())
	}

	return
}

// PhoneExists проверяет наличие номера телефона в таблице address_book.
// Возвращает ошибку "phone number already in use", если номер телефона уже используется.
// Возвращает nil, если номер телефона не найден.
//
// Пример использования:
//
//		err := psg.PhoneExists("+71234567890")
//		if err != nil {
//		    fmt.Println(err.Error()) // "phone number already in use"
//		} else {
//	     fmt.Println("phone number not found")
//		}
func (p *Psg) PhoneExists(phone string) error {
	wErr, err := pkg.NewWrappedErrorWithFile("(p *Psg) PhoneExists()")
	if err != nil {
		log.Println("(p *Psg) PhoneExists(): NewWrappedErrorWithFile()", err)
	}

	sqlCommand := `SELECT phone FROM address_book WHERE phone = $1`
	row := p.conn.QueryRow(context.Background(), sqlCommand, phone)

	var existingPhone string
	err = row.Scan(&existingPhone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		wErr.Specify(err, "row.Scan(&existingPhone)").LogError()
		return err
	}
	if existingPhone == phone {
		err = errors.New("phone number already in use")
		wErr.LogMsg(err.Error())
		return err
	}

	return errors.New("phone number already in use")
}
