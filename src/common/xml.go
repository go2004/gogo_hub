/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-5-5
 * Time: 下午3:05
 * To change this template use File | Settings | File Templates.
 */
package common

/*
import (
	"encoding/xml"
	"log"
	"io/ioutil"
)

type Result struct {
	XMLName xml.Name `xml:"persons"`
	Persons []Person `xml:"person"`
}
type Person struct {
	Name      string `xml:"name,attr"`
	Age       int `xml:"age,attr"`
	Career    string `xml:"career"`
	Interests []string `xml:"interests>interest"`
}

func main() {
	content, err := ioutil.ReadFile("studygolang.xml")
	if err != nil {
		log.Fatal(err)
	}
	var result Result
	err = xml.Unmarshal(content, &result)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)
	persons := result.Persons[0]
	log.Println(len(persons.Interests))
}
*/
