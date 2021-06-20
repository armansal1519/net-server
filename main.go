package main

import (
	"encoding/json"
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"io/ioutil"
	"log"
	"os"
	randomcolor "server/randomColor"
)

// MessageObject Basic chat message object

type conResp struct {
	Action string `json:"action"`
	UserName	string `json:"user_name"`
	RoomName string `json:"room_name"`
	Color string `json:"color"`
	Uuid  string `json:"uuid"`
}

type msgResp struct {
	Action string `json:"action"`
	RoomName string `json:"room_name"`
	UserMsg string `json:"user_msg"`
	Color string `json:"color"`
	Name string `json:"name"`
	Link string `json:"link"`
	FileColor string `json:"file_color"`
	FileName string	`json:"file_name"`
}

type room struct {
	RomeName string
	Clients  []client
}
type client struct {
	id   string
	name string
	color string
}

func remove(slice []client, s int) []client {
	return append(slice[:s], slice[s+1:]...)
}

var rooms []room



func main() {
	fileColor := make(map[string]string)

	// The key for the map is message.to

	// Start a new Fiber application
	app := fiber.New()
	app.Use(cors.New())

	// Setup the middleware to retrieve the data sent in first GET request
	//app.Use(func(c *fiber.Ctx) error {
	//	// IsWebSocketUpgrade returns true if the client
	//	// requested upgrade to the WebSocket protocol.
	//	if websocket.IsWebSocketUpgrade(c) {
	//		c.Locals("allowed", true)
	//		return c.Next()
	//	}
	//	return fiber.ErrUpgradeRequired
	//})

	// Multiple event handling supported
	//ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
	//	//fmt.Println(fmt.Sprintf("Connection event 1 - User: %s", ep.Kws))
	//	fmt.Println(clients)
	//})
	//
	//
	// On message event
	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {

		roomName:= ep.Kws.GetStringAttribute("room")
		//fmt.Println(ep.Data)
		//message := MessageObject{}

		type data struct {
			Msg      string  `json:"msg"`
			FileName string  `json:"name"`
			Arr      []uint8 `json:"arr"`
		}
		m := data{}


		err := json.Unmarshal(ep.Data, &m)
		if err != nil {
			fmt.Println(1, err)
			return
		}
		format:= nameToFormat(m.FileName)
		color,ok:=fileColor[format]
		if !ok {
			rc:=randomcolor.GetRandomColorInHex()
			fileColor[format]=rc
			color=rc
		}

		//fmt.Println(m)
		c:=findUserByUuid(ep.Kws.UUID)
		if c.id=="" {
			log.Fatal("uuid not found")
		}



		if m.Msg!="" &&( m.Arr==nil || len(m.Arr)<2) {
			resp:=msgResp{
				Action:   "newMsg",
				RoomName: roomName,
				UserMsg:  m.Msg,
				Color: c.color,
				Name: c.name,

			}
			b,err:=json.Marshal(resp)
			if err != nil {
				log.Fatal(err)
			}
			err=emitToRoom(roomName,b,ep.Kws,ep.Kws.UUID)

		} else {
			p:=fmt.Sprintf("/%v/%v", roomName,m.FileName)
			err = ioutil.WriteFile(fmt.Sprintf(".%v",p), m.Arr, 0644)
			if err != nil {
				log.Fatal(err)
			}
			resp:=msgResp{
				Action:   "newMsg",
				RoomName: roomName,
				UserMsg:  m.Msg,
				Color: c.color,
				Name: c.name,
				Link: p,
				FileColor: color,
				FileName: m.FileName,
			}
			b,err:=json.Marshal(resp)
			if err != nil {
				log.Fatal(err)
			}
			err=emitToRoom(roomName,b,ep.Kws,ep.Kws.UUID)

		}

	})

	//
	// On disconnect event
	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		//delete(clients, ep.Kws.GetStringAttribute("user_id"))
		fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.UUID))
		//fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("room")))
		for i, _ := range rooms {
			if rooms[i].RomeName == ep.Kws.GetStringAttribute("room") {
				for j, u := range rooms[i].Clients {
					if u.id == ep.Kws.UUID {
						rooms[i].Clients = remove(rooms[i].Clients, j)
					}
				}
			}
		}
		fmt.Println(rooms)
	})

	// On close event
	// This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventClose, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		//delete(clients, ep.Kws.GetStringAttribute("user_id"))
		fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.UUID))
		//fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("room")))
		for i, _ := range rooms {
			if rooms[i].RomeName == ep.Kws.GetStringAttribute("room") {
				for j, u := range rooms[i].Clients {
					if u.id == ep.Kws.UUID {
						rooms[i].Clients = remove(rooms[i].Clients, j)
					}
				}
			}
		}
		fmt.Println(rooms)
	})
	//
	// On error event
	//ikisocket.On(ikisocket.EventError, func(ep *ikisocket.EventPayload) {
	//	fmt.Println(fmt.Sprintf("Error event - User: "))
	//})

	//ikisocket.On("abc", func(ep *ikisocket.EventPayload) {
	//	fmt.Println("hi")
	//})
	////
	app.Get("/ws/:room/:userName", ikisocket.New(func(kws *ikisocket.Websocket) {

		// Retrieve the user id from endpoint
		roomName := kws.Params("room")
		userName := kws.Params("userName")
		rc:=randomcolor.GetRandomColorInHex()

		// Add the connection to the list of the connected clients
		// The UUID is generated randomly and is the key that allow
		// ikisocket to manage Emit/EmitTo/Broadcast
		path := fmt.Sprintf("./%v", roomName)
		//if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println(path)
		for _, r := range rooms {
			if r.RomeName == roomName {
				for _, c := range r.Clients {
					resp:=conResp{
						Action: "addUser",
						UserName: c.name,
						RoomName: roomName,
						Color: c.color,
						Uuid:     c.id,
					}

					b, err := json.Marshal(resp)
					if err != nil {
						fmt.Println(err)
					}
					err=kws.EmitTo(kws.UUID,b)
					if err!=nil {
						fmt.Println(err)
					}
				}
			}
		}



		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Println("exist")
		}

		doesRoomExist := false
		for i, _ := range rooms {
			if rooms[i].RomeName == roomName {
				c := client{
					id:   kws.UUID,
					name: userName,
					color: rc,
				}
				rooms[i].Clients = append(rooms[i].Clients, c)
				doesRoomExist = true
				break
			}
		}
		if !doesRoomExist {
			c := client{
				id:   kws.UUID,
				name: userName,
				color: rc,

			}
			r := room{
				RomeName: roomName,
				Clients:  []client{c},
			}
			rooms = append(rooms, r)
		}

		//clients[userId] = kws.UUID
		
		resp:=conResp{
			Action: "addUser",
			UserName: userName,
			RoomName: roomName,
			Color: rc,
			Uuid:     kws.UUID,
		}

		b, err := json.Marshal(resp)
		//fmt.Println(r,string(b))

		// Every websocket connection has an optional session key => value storage
		kws.SetAttribute("room", roomName)


		err = emitToRoom(roomName, b, kws,kws.UUID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(rooms)
		//Broadcast to all the connected users the newcomer
		//kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", roomName, kws.UUID)), true)
		////Write welcome message
		//kws.Emit([]byte(fmt.Sprintf("Hello user: %s with UUID: %s", roomName, kws.UUID)))
	}))

	//app.Post("ws",ikisocket.New(func(kws *ikisocket.Websocket) {
	//
	//	// Retrieve the user id from endpoint
	//	userId := kws.Params("id")
	//	// Add the connection to the list of the connected clients
	//	// The UUID is generated randomly and is the key that allow
	//	// ikisocket to manage Emit/EmitTo/Broadcast
	//	clients[userId] = kws.UUID
	//
	//	// Every websocket connection has an optional session key => value storage
	//	kws.SetAttribute("user_id", userId)
	//
	//	//Broadcast to all the connected users the newcomer
	//	kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true)
	//	//Write welcome message
	//	kws.Emit([]byte(fmt.Sprintf("Hello user: %s with UUID: %s", userId, kws.UUID)))
	//}))

	app.Get("/rooms", func(c *fiber.Ctx) error {
		roomArr:=make([]string,0)
		for _, r := range rooms {
			roomArr=append(roomArr,r.RomeName)
		}
		return c.JSON(fiber.Map{"rooms":roomArr})
	})
	app.Get("/:room/:file", func(c *fiber.Ctx) error {
		r:=c.Params("room")
		f:=c.Params("file")
		return c.Download(fmt.Sprintf("./%v/%v",r,f))
		//return c.SendFile(fmt.Sprintf("./%v/%v",r,f));
		// => Download report-12345.pdf
		// => Download report.pdf
	})

	log.Fatal(app.Listen(":3003"))
}

func emitToRoom(roomName string, msg []byte, kws *ikisocket.Websocket,uuid string) error {
	selectedRoom := room{}
	for i, _ := range rooms {
		if roomName == rooms[i].RomeName {
			selectedRoom = rooms[i]
		}
	}

	if len(selectedRoom.Clients) > 0 {
		var uuids []string
		for _, v := range selectedRoom.Clients {
			if v.id != uuid {
				uuids = append(uuids, v.id)
			}
		}

		for _, uuid := range uuids {
			err := kws.EmitTo(uuid, msg)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

func findUserByUuid(uuid string) client {
	for _, r := range rooms {
		for _, c := range r.Clients {
			if c.id==uuid{
				return c
			}
		}
	}
	return client{}
}

func nameToFormat(name string) string{
	runes:=[]rune(name)
	dot:=rune('.')
	format:=make([]rune,0)
	for i := len(runes)-1; i >0 ; i-- {
		if runes[i]==dot {
			break
		}
		format=append(format,runes[i])
	}





	return string(format)
}