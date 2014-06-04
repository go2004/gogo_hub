package main

import (
	"database/sql"    //这包一定要引用
	"fmt"    //这个前面一章讲过
	_"mysql"
	//	"strconv" //这个是为了把int转换为string
	"log"
	"time"
	//	"os"
	"common"
)

type DbMysql struct {
	db *sql.DB
	url          string
	state        bool        //状态 true:正常，false:关闭
	opFlag       chan uint8  //退出标记，0:正常,1：退出, 2:PING
	querychannel chan string //sql缓存
}

/* 初始化数据库引擎 */
func DbInit() (*DbMysql, error) {
	hostIP := common.GetElement("MySQL", "HostIP", "127.0.0.1");
	port := common.GetElement("MySQL", "Port", "3306");
	user := common.GetElement("MySQL", "User", "root");
	pwd := common.GetElement("MySQL", "Pwd", "1234");
	dbName := common.GetElement("MySQL", "DbName", "");

	DbConn, err := dbInit(hostIP, port, user, pwd, dbName);
	return DbConn, err
}

/* 初始化数据库引擎 */
func dbInit(hostIP, port, user, pwd, dbname string) (*DbMysql, error) {
	mydb := new(DbMysql);
	url := user + ":" + pwd + "@tcp(" + hostIP + ":" + port + ")/" + dbname + "?charset=utf8"
	db, err := sql.Open("mysql", url);
	//db,err := sql.Open("mysql","root:123@tcp(127.0.0.1:3306)/test?charset=utf8");
	//第一个参数 ： 数据库引擎
	//第二个参数 : 数据库DSN配置。Go中没有统一DSN,都是数据库引擎自己定义的，因此不同引擎可能配置不同
	//本次演示采用http://code.google.com/p/go-mysql-driver
	if err != nil {
		log.Println("database initialize error : ", err.Error());
		return nil, err;
	}
	mydb.db = db;
	mydb.url = url;
	mydb.state = true;
	mydb.opFlag = make(chan uint8);
	mydb.querychannel = make(chan string, SqlChannelSize);
	//自检功能
	go DbPing(mydb);

	log.Println("database initialize succes.... ");
	return mydb, nil
}
func DbClose(mydb *DbMysql) {
	mydb.db.Close()
}

func DbPing(mydb *DbMysql) {
	var opFlag uint8;
	var tick uint = 0;
	var query string;
	for {
		select {
		case opFlag = <-mydb.opFlag:              //1：退出
			if (opFlag == 1) {    break; }
		default:
		}

		time.Sleep(DbUpdateQuerychannel);
		if !mydb.state {     //状态不正确，重新连接
			db, err := sql.Open("mysql", mydb.url);
			if err == nil {
				mydb.db = db;
				mydb.state = true;
				continue;
			}
			log.Println("database initialize error : ", err.Error());
			continue;
		}

		if (tick > DbUpdatePingMaxTick) {
			if (mydb.db.Ping() != nil) {   //ping检测mysql
				mydb.state = false;
				mydb.db.Close();
			}
			tick = 0;
		}
		tick += 1;

		//从缓存中执行
		for moreData := true; moreData; {
			select {
			case query = <-mydb.querychannel:
				mydb.Exec(query);
				tick = 0
			default:
				moreData = false
			}
		}
	}
	log.Println("database is quit......");
}

//Exec executes a query  returning any rows.
func (mydb *DbMysql) Prepare(query string) {
	if mydb.db == nil || !mydb.state {    //conn is close,push to lists
		return;
	}
	stmt, err := mydb.db.Prepare(query);
	if err != nil {
		log.Println(err.Error());
		return;
	}
	defer stmt.Close();
	if result, err := stmt.Exec(); err == nil {
		if c, err := result.RowsAffected(); err == nil {
			log.Println(query, ", update count : ", c);
		}
	}else {
		log.Println("err = ", err, "\n", result)
	}
}

//Exec executes a query without returning any rows.
func (mydb *DbMysql) Exec(query string) {
	if mydb.db == nil || !mydb.state {    //conn is close,push to lists
		return;
	}
	_, err := mydb.db.Exec(query);
	if err != nil {
		log.Println("Query = ", query, " : ", err.Error());
		return;
	}
}

//查询实例
func (mydb *DbMysql) Query(query string) {
	if mydb.db == nil || !mydb.state {
		return;
	}
	rows, err := mydb.db.Query(query);
	if err != nil {
		log.Println("Query = ", query, " : ", err.Error());
		return;
	}
	defer rows.Close();

	var id int;
	var name string;
	var age int;
	for rows.Next() {
		if err := rows.Scan(&id, &name, &age); err == nil {
			fmt.Print(id);
			fmt.Print(name);
			fmt.Print(age);
		}
	}
}

/*使用实例
 mydb, err := DbInit("root", "1234", "127.0.0.1", "3306", "test");
if (err == nil) {
mydb.Exec("call p3(2)");
opFlag := <-mydb.opFlag;
}
  */
