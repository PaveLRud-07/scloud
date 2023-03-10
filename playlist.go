package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

//структура трека

type sound struct {
	Name     string
	Duration int
}

// создал структуру для хранения  плейлиста
type playlist []sound

// сканируем элементы управления (Текстовые команды)
func sscan(a string, prev *string) string {
	//время ожидания выполнения горутинны (если поставить меньше не успеваешь вводить)
	to := time.After(5 * time.Second)
	a = *prev
	go fmt.Scan(&a)
	for {
		select {
		case <-to:
			*prev = a
			return a
		}
	}
}

// в зависимоти от введёной команды получаем результат работы
func buttons(a string, arr *playlist, c chan bool, e chan string) {
	switch a {
	case "pause":
		fmt.Println("pause")
		fmt.Print(".")
		fmt.Print("..")
		fmt.Print("..")
		time.Sleep(5 * time.Second)
	case "add":
		*arr = arr.AddSong()
	case "next":
		fmt.Println("Пропускаем трек")
		c <- true
	case "play":
		fmt.Print("play")
	case "exit":
		writeMusic(*arr)
		os.Exit(0)
	default:
		fmt.Print(".")
	}
}

// имитируем проигрывание трека
// в случае если получаем команду пауза
// ждём пока на вход подадут другую команду
func Play(a sound, e chan string, check chan bool) {
	fmt.Print("playing ", a.Name)
	for i := 0; time.Second*time.Duration(i) < time.Second*time.Duration(a.Duration); {
		for {
			if <-e != "pause" {
				time.Sleep(2 * time.Second)
				fmt.Printf(".")
				time.Sleep(2 * time.Second)
				fmt.Printf("..")
				time.Sleep(2 * time.Second)
				fmt.Printf("...\n")
				i = i + 1
				//проверка на время воспроизвидения трека
				if time.Second*time.Duration(i) <= time.Second*time.Duration(a.Duration) {
					break
				}
			}
		}
		if time.Second*time.Duration(i) <= time.Second*time.Duration(a.Duration) {
			check <- true
		}
	}
}

// создаём горутину что бы проерять введёную команду или её отсутствие
func chekButton(e chan string, a, p *string, arr playlist, c chan bool) {
	for {
		//получаем текстовую команду
		*a = sscan(*a, *&p)

		//проверяем введёную команду
		buttons(*a, &arr, c, e)
		//блок исключений
		if *a != "pause" {
			e <- "a"
			return
		}
		e <- *a
	}
}

// проигрываем плей лист
func PlayAll(arr playlist, e chan string) {
	fmt.Println("Для управления 'плеером' Необходимо вводить команды")
	fmt.Println("play - для проигрывания трека")
	fmt.Println("pause - дял паузы")
	fmt.Println("add - для добавления трека в конец")
	fmt.Println("next - для следующего трека")
	fmt.Println("exit - для выхода")
	var a, p string
	check := make(chan bool)
	for _, v := range arr {
		go Play(v, e, check)
		go chekButton(e, &a, &p, arr, check)
		//если в канал чек переданно значение
		//трек закончил воспроизведение и начинается воспроизвидение следующего трека
		<-check
	}
}

/* Переделки и недоделки
Функцию следующего трека надо переделать.
сделав канал чек пренимающим значения типа инт
после чего в цикле плейлиста присвоить переменной i
значение возвращаемое каналом чек.
В сам канал в зависимости от функции next or prev передаются значения i+1,i-1

починить addsong

написать функцию которая переводит входящие сообщения от пользователя в нижний регистр

переписать функцию sscan на цикл фор пока не будет закрыт канал
*/
// Добавляем трек в конец (не работает)
func (a *playlist) AddSong() playlist {
	var b sound
	fmt.Println("Введите название трека и его дляительность")
	fmt.Scan(&b.Name, &b.Duration)
	*a = append(*a, b)
	return *a
}

// загружаем из json файла информацию о треке
// и записываем их в массив стрктур
func loadMusik() []sound {
	fileName, err := os.Open("music.json")
	if err != nil {
		log.Fatal(err)
	}

	defer fileName.Close()

	data, err := ioutil.ReadAll(fileName)

	if err != nil {
		log.Fatal(err)
	}

	var result []sound

	jsonErr := json.Unmarshal(data, &result)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result
}

// записываем в json файл новый плейлист
func writeMusic(arr []sound) {
	json_data, err := json.Marshal(arr)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile("music.json", os.O_RDWR|os.O_TRUNC, 0775)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.Write(json_data)
}

func main() {
	e := make(chan string)
	arr := loadMusik()
	PlayAll(arr, e)
}
