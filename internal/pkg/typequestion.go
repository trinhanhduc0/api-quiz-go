package pkg

import (
	utils "my_app/internal/util"

	"go.mongodb.org/mongo-driver/bson"
)

// TypeQuestion trả về tên trường cho danh sách và câu trả lời dựa trên loại câu hỏi
func TypeQuestion(typeQuestion string) (string, string) {

	switch typeQuestion {
	case "fill_in_the_blank":
		return "fill_in_the_blank", "correct_answer"

	case "single_choice_question", "multiple_choice_question":
		return "options", "iscorrect"

	case "match_choice_question":
		return "options", "match"

	case "order_question":
		return "options", "order"

	default:
		return "", ""
	}
}

func ProcessQuestion(question bson.M) {
	switch question["type"].(string) {
	case "match_choice_question":
		return
	default:
		arrayField, answerField := TypeQuestion(question["type"].(string))

		if arrayField == "" || answerField == "" {
			return // Bỏ qua nếu không xác định được các trường cần thiết
		}

		if array, ok := question[arrayField].(bson.A); ok {
			convertedArray := utils.ConvertToBsonMArray(array)
			utils.RemoveAnswer(convertedArray, answerField)
			question[arrayField] = convertedArray
		} else if array, ok := question[arrayField].([]bson.M); ok {
			utils.RemoveAnswer(array, answerField)
			question[arrayField] = array
		}
	}
}
