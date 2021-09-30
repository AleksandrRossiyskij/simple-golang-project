package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type DataBase struct {
	mutex   sync.Mutex
	mutexes map[string]*sync.Mutex
	dir     string
	name    string
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

/* DataBase Implementation */

func New(dir, name string) (*DataBase, error) {
	dir = filepath.Clean(dir)

	baseOfData := DataBase{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		name:    name,
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if name == f.Name() {
			fmt.Fprintf(os.Stderr, "Using '%s' (database already existrs)\n", dir)
			return &baseOfData, nil
		}
	}

	fmt.Printf("Creating the database at '%s'...\n", dir)
	return &baseOfData, os.MkdirAll(dir, 0755)
}

func (d *DataBase) Write(resource string, v interface{}) error {
	if d.name == "" {
		return fmt.Errorf("Mission collection - no place to save record!")
	}

	if resource == "" {
		return fmt.Errorf("Missing resource - unable to save record!")
	}

	mutex := d.createMutex()
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, d.name)
	fanalPath := filepath.Join(dir, resource+".json")
	tmpPath := fanalPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return nil
	}

	return os.Rename(tmpPath, fanalPath)
}

func (d *DataBase) Read(resource string, v interface{}) error {
	if d.name == "" {
		return fmt.Errorf("Mission collection - unable to read!")
	}

	if resource == "" {
		return fmt.Errorf("Missing resource - unable to read record")
	}
	record := filepath.Join(d.dir, d.name, resource)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(record + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *DataBase) ReadAll() ([]string, error) {
	if d.name == "" {
		return nil, fmt.Errorf("Mission collection - unable to read!")
	}

	dir := filepath.Join(d.dir, d.name)
	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := ioutil.ReadDir(dir)

	var records []string

	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		records = append(records, string(b))
	}
	return records, nil
}

func (d *DataBase) Delete(resource string) error {
	path := filepath.Join(d.name, resource)
	mutex := d.createMutex()
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("Unable to find file or directory name %v\n", path)
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)
	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}

	return nil
}

func (d *DataBase) createMutex() *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	mu, ok := d.mutexes[d.name]

	if !ok {
		mu = &sync.Mutex{}
		d.mutexes[d.name] = mu
	}

	return mu
}

func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return
}

/*
 POST request func
*/
func form(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/form" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		name := r.FormValue("Name")
		age := json.Number(r.FormValue("Age"))
		contact := r.FormValue("Contact")
		company := r.FormValue("Company")
		city := r.FormValue("City")
		state := r.FormValue("State")
		country := r.FormValue("Country")
		pincode := json.Number(r.FormValue("Pincode"))

		user := User{
			name,
			age,
			contact,
			company,
			Address{
				city,
				state,
				country,
				pincode},
		}

		db.Write(user.Name, User{
			Name:    user.Name,
			Age:     user.Age,
			Contact: user.Contact,
			Company: user.Company,
			Address: user.Address,
		})

		dir, _ := os.Getwd()
		tmp, err := template.ParseFiles(dir + "/templates/userTemplate.html")
		if err != nil {
			log.Fatal(err)
		}
		tmp.Execute(w, user)
	default:
		fmt.Fprintf(w, "Sorry, only POST methods are supported.")
	}
}

var (
	directory = "./"
	db, err   = New(directory, "users")
)

func main() {
	fileServer := http.FileServer(http.Dir("./public"))
	http.Handle("/", fileServer)

	http.HandleFunc("/form", form)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error %v", err)
	}
	/*
		employees := []User{
			{"John", "23", "5 (535) 45-43-561", "MyTech", Address{"Wien", "Vienna", "Austria", "453"}},
			{"Hans", "20", "5 (535) 35-44-561", "MyTech", Address{"Wien", "Vienna", "Austria", "345"}},
			{"Peter", "18", "5 (435) 35-39-565", "MyTech", Address{"Wien", "Vienna", "Austria", "389"}},
			{"Henry", "34", "5 (535) 33-87-541", "MyTech", Address{"Wien", "Vienna", "Austria", "325"}},
			{"Neo", "19", "1 (545) 24-87-541", "MyTech", Address{"San Franciscor", "California", "USA", ""}},
			{"Robert", "26", "5 (535) 38-85-541", "MyTech", Address{"Wien", "Vienna", "Austria", "428"}},
		}

		for _, value := range employees {
			db.Write(value.Name, User{
				Name:    value.Name,
				Age:     value.Age,
				Contact: value.Contact,
				Company: value.Company,
				Address: value.Address,
			})
		}
	*/ // Write

	/*
		aUser := User{}
		db.Read("Hans", &aUser)

		fmt.Printf("Name: %s\n", aUser.Name)
	*/ // Read one file

	/*
		records, err := db.ReadAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error %v", err)
		}
		fmt.Println(records)
	*/ //ReadAll

	/*
		if err := db.Delete("John"); err != nil {
			fmt.Fprintf(os.Stderr, "Error", err)
		}
	*/ // Delete John.json

	/*
		if err := db.Delete(""); err != nil {
			fmt.Fprintf(os.Stderr, "Error", err)
		}
	*/ // DeleteAll

	fmt.Printf("Starting server at port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
