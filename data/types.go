package data

type MessageDTO struct {
	Author  string
	Content string
}

type HotWordDTO struct {
	Word      string
	Frequency int
}

//func (message *pb.ChatStream_Message_) MarshallBinary(data interface{}) {
//
//}
