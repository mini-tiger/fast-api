package sqlBuild

import (
	"fmt"

	"github.com/mini-tiger/fast-api/dbManager"
	"gorm.io/gorm"
)

type MysqlType struct {
	tx            *gorm.DB
	orderByString string
	groupByFiled  string
	filed         string
	sql           string
	limit         string
	tableName     string
	joinTableList []string
	whereList     []string
	withCount     string
}

func Create() *MysqlType {
	return &MysqlType{
		filed: "*",
	}
}

func (m *MysqlType) OpenTx(tx *gorm.DB) *MysqlType {
	m.tx = tx
	return m
}

func (m *MysqlType) db() *gorm.DB {
	if nil != m.tx {
		return m.tx
	}
	return nil
}

func (m *MysqlType) SetPage(page, pageSize int) *MysqlType {
	if 0 >= page {
		page = 1
	}
	start := (page - 1) * pageSize
	m.limit = fmt.Sprintf("limit %d, %d", start, pageSize)
	return m
}

func (m *MysqlType) SetTableName(tableName string) *MysqlType {
	m.tableName = tableName
	return m
}

func (m *MysqlType) SetFiled(filed string) *MysqlType {
	m.filed = filed
	return m
}

func (m *MysqlType) SetJoinTable(joinTable string) *MysqlType {
	m.joinTableList = append(m.joinTableList, joinTable)
	return m
}

func (m *MysqlType) SetGroupBy(groupByFiled string) *MysqlType {
	m.groupByFiled = groupByFiled
	return m
}

func (m *MysqlType) SetOrderBy(orderByString string) *MysqlType {
	m.orderByString = orderByString
	return m
}

func (m *MysqlType) Where(field string, op string, value interface{}) *MysqlType {
	var condition string

	switch op {
	case "=", "!=":
		switch v := value.(type) {
		case string:
			condition = fmt.Sprintf("%s %s '%s'", field, op, v)
		default:
			condition = fmt.Sprintf("%s %s %v", field, op, v)
		}
	case "in", "not in":
		var inVals string
		switch v := value.(type) {
		case []string:
			for i, s := range v {
				inVals += fmt.Sprintf("'%s'", s)
				if i < len(v)-1 {
					inVals += ", "
				}
			}
		case []int:
			for i, num := range v {
				inVals += fmt.Sprintf("%d", num)
				if i < len(v)-1 {
					inVals += ", "
				}
			}
		case []interface{}:
			for i, elem := range v {
				switch el := elem.(type) {
				case string:
					inVals += fmt.Sprintf("'%s'", el)
				default:
					inVals += fmt.Sprintf("%v", el)
				}
				if i < len(v)-1 {
					inVals += ", "
				}
			}
		default:
			// support single value fallback
			switch v1 := v.(type) {
			case string:
				inVals = fmt.Sprintf("'%s'", v1)
			default:
				inVals = fmt.Sprintf("%v", v1)
			}
		}
		condition = fmt.Sprintf("%s %s (%s)", field, op, inVals)
	case "like":
		switch v := value.(type) {
		case string:
			condition = fmt.Sprintf("%s LIKE '%s'", field, v)
		default:
			condition = fmt.Sprintf("%s LIKE '%v'", field, v)
		}
	default:
		// unsupported op, raw fallback
		condition = fmt.Sprintf("%s %s %v", field, op, value)
	}

	m.whereList = append(m.whereList, condition)
	return m
}

func (m *MysqlType) WhereRaw(where string) *MysqlType {
	m.whereList = append(m.whereList, where)
	return m
}

func (m *MysqlType) Get(data interface{}) {
	m.build()
	db := m.db()
	if nil == db {
		db = dbManager.GetInstance()
	}
	db.Raw(m.sql).Scan(data)
}

func (m *MysqlType) GetWithCount(data interface{}) int {
	m.withCount = "SQL_CALC_FOUND_ROWS"

	m.build()
	// 定一个临时结构体用于获取总条数
	total := struct{ Total int }{}
	// 如果开启了外部事物，则这里无需开启事务
	db := m.db()
	autoTx := false
	if nil == db {
		autoTx = true
		db = dbManager.GetInstance().Begin()
	}
	db.Raw(m.sql).Scan(data)
	db.Raw("SELECT FOUND_ROWS() as total").Scan(&total)

	// 没有外部事物情况下，自动事务要提交
	if autoTx {
		db.Commit()
	}
	return total.Total
}

func (m *MysqlType) build() {
	// 处理join语句
	for _, joinTable := range m.joinTableList {
		m.sql = fmt.Sprintf("%s %s", m.sql, joinTable)
	}

	// 处理where条件
	if 0 < len(m.whereList) {
		m.sql = fmt.Sprintf("%s where", m.sql)
	}
	for index, where := range m.whereList {
		if 0 == index {
			m.sql = fmt.Sprintf("%s %s", m.sql, where)
		} else {
			m.sql = fmt.Sprintf("%s and %s", m.sql, where)
		}
	}

	// 处理groupBy
	if "" != m.groupByFiled {
		m.sql = fmt.Sprintf("%s group by %s", m.sql, m.groupByFiled)
	}

	// 处理orderBy
	if "" != m.orderByString {
		m.sql = fmt.Sprintf("%s order by %s", m.sql, m.orderByString)
	}

	if "" != m.limit {
		// 处理limit
		m.sql = fmt.Sprintf("%s %s", m.sql, m.limit)
	}

	// 处理sql开始部分
	m.sql = fmt.Sprintf("select %s %s from %s a %s", m.withCount, m.filed, m.tableName, m.sql)
}
