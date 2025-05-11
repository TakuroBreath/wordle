package main

import (
	"fmt"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

func main() {
	// Токен бота из конфигурации
	botToken := "7613582783:AAE5_-7beHhGHQi60PIPAfddMGecE_Dnmq0"

	// Тестовые данные инициализации
	initData := "user=%7B%22id%22%3A656332640%2C%22first_name%22%3A%22Max%20Takuro%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22capybaraenjoy%22%2C%22language_code%22%3A%22en%22%2C%22is_premium%22%3Atrue%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2FveTetCZa8ASYFqG0G6tVVEffbiY5d14-Kzg-VYsi5JI.svg%22%7D&chat_instance=2121117271269367483&chat_type=sender&auth_date=1746965729&signature=zlRKwXlbFMHfZkY8ICrZ-UwXyjNAGgb36rY8uA4LS9vp-2j0mFg-jjoaPvPIo-fWQH8r5bynDk94Pvt2kjDuCQ&hash=029f759894dcaebd04fac070ba1d9b0e5b164505b11b490c767c54eb039b7901"

	// Парсим данные
	data, err := initdata.Parse(initData)
	if err != nil {
		fmt.Printf("Ошибка парсинга данных: %v\n", err)
		return
	}

	// Выводим информацию о пользователе
	fmt.Printf("ID пользователя: %d\n", data.User.ID)
	fmt.Printf("Имя пользователя: %s\n", data.User.Username)
	fmt.Printf("Имя: %s\n", data.User.FirstName)
	fmt.Printf("Фамилия: %s\n", data.User.LastName)

	// Валидируем данные
	err = initdata.Validate(initData, botToken, time.Hour)
	if err != nil {
		fmt.Printf("Ошибка валидации: %v\n", err)
		return
	}

	fmt.Println("Валидация успешна!")
}
