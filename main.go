package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gameparrot/mcpelauncher_pcproxy/auth"
	"github.com/gameparrot/mcpelauncher_pcproxy/proxy"

	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var xsts *auth.XBLToken
var lastUpdateTime time.Time

var currentAddr string
var xuid string

func updateXstsToken() (*auth.XblTokenObtainer, error) {
	if time.Since(lastUpdateTime) < 45*time.Minute {
		return nil, nil
	}
	liveToken, err := Auth.Token()
	if err != nil {
		return nil, fmt.Errorf("request Live Connect token: %w", err)
	}

	obtainer, err := auth.NewXblTokenObtainer(liveToken, context.Background())
	if err != nil {
		return nil, fmt.Errorf("request Live Device token: %w", err)
	}

	xsts, err = obtainer.RequestXBLToken(context.Background(), "https://multiplayer.minecraft.net/")
	if err != nil {
		return nil, fmt.Errorf("request XBOX Live token: %w", err)
	}
	lastUpdateTime = time.Now()
	return obtainer, nil
}

func login() error {
	obtainer, err := updateXstsToken()
	if err != nil {
		return err
	}
	accountInfoTok, err := obtainer.RequestXBLToken(context.Background(), "http://xboxlive.com")
	if err != nil {
		return fmt.Errorf("request XBOX site token: %w", err)
	}
	xuid = accountInfoTok.AuthorizationToken.DisplayClaims.UserInfo[0].XUID
	return nil
}

func main() {
	Auth.Startup()
	if !Auth.LoggedIn() {
		Auth.Login(context.Background(), &DeviceTypeIOSPreview)
	}

	if err := login(); err != nil {
		panic(err)
	}

	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter server ip: ")
	reader.Scan()
	currentAddr = reader.Text()

	ip, port, err := net.SplitHostPort(currentAddr)
	if err != nil {
		ip = currentAddr
		port = "19132"
	}
	bestIp := GetLowestPingIP(ip, port)
	fmt.Println("Found best IP: " + bestIp)
	currentAddr = net.JoinHostPort(bestIp, port)

	l, err := raknet.Listen("127.0.0.1:19132")
	if err != nil {
		panic(err)
	}

	fmt.Println("Connect to 127.0.0.1 port 19132")

	t := time.NewTicker(3 * time.Second)
	go func() {
		for {
			<-t.C
			p, err := raknet.Ping(currentAddr)
			if err == nil {
				l.PongData(p)
			}
		}
	}()

	for {
		conn, err := l.Accept()
		if err == nil {
			go handleConn(conn.(*raknet.Conn))
		}
	}

}

func handleConn(conn *raknet.Conn) {
	fmt.Println("New connection: " + conn.RemoteAddr().String())
	clientConn := proxy.NewProxyConn(conn, false)
	clientConn.SetAuthEnabled(true)
	clientConn.SetSaltAndKey()
	if err := clientConn.ReadLoop(); err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	if clientConn.IdentityData().XUID != xuid {
		clientConn.WritePacket(&packet.Disconnect{Message: "Error: Account mismatch"})
		time.Sleep(100 * time.Millisecond)
		conn.Close()
		return
	}

	fmt.Println("Client logged in")

	rkConn, err := raknet.Dial(currentAddr)
	if err != nil {
		fmt.Println(err)
		conn.Close()
	}

	defer rkConn.Close()
	defer conn.Close()

	serverConn := proxy.NewProxyConn(rkConn, false)

	cd := clientConn.ClientData()
	cd.DeviceOS = 7
	cd.PlatformType = 0
	cd.GameVersion += ".24"
	cd.DefaultInputMode = 1
	cd.ServerAddress = currentAddr
	cd.DeviceModel = "MCPELAUNCHER PCPROXY"

	serverConn.Login(cd, xsts, clientConn.Protocol())
	fmt.Println("Server logged in")

	go func() {
		defer rkConn.Close()
		defer conn.Close()
		for {
			pk, err := clientConn.ReadCompressedPacket()
			if err != nil {
				fmt.Println(fmt.Errorf("read client packet: %w", err))
				return
			}
			if _, err := serverConn.WriteCompressedPacket(append([]byte{254}, pk...)); err != nil {
				fmt.Println(err)
				return
			}
		}
	}()

	for {
		pk, err := serverConn.ReadCompressedPacket()
		if err != nil {
			fmt.Println(fmt.Errorf("read server packet: %w", err))
			return
		}
		if _, err := clientConn.WriteCompressedPacket(append([]byte{254}, pk...)); err != nil {
			fmt.Println(err)
			return
		}
	}
}
