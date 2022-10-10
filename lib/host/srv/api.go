package srv

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

func listenApi(conn *net.TCPConn) {
	buf := make([]byte, 8192)
	i, err := conn.Read(buf)
	raw := buf[:i]
	if err != nil {
		log.Print("Error in reading message from api: ", err)
		conn.Close()
		return
	}
	request := common.ApiRequestData{}
	err = json.Unmarshal(raw, &request)
	if err != nil {
		log.Print("Error in unmarshaling message from api: ", err)
		conn.Close()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		io.Copy(io.Discard, conn)
		cancel()
	}()
	err = apiExec(&request, ctx)
	responce := common.NewApiResultData("", 0)
	if err != nil {
		responce.Message = err.Error()
		responce.Code = 1
	}
	raw, err = json.Marshal(responce)
	if err != nil {
		log.Print("Error in marshaling message to api: ", err)
		conn.Close()
		return
	}
	_, err = conn.Write(raw)
	if err != nil {
		log.Print("Error in writing responce to api: ", err)
	}
	conn.Close()
}

var ErrPackageNotFound = errors.New("invalid plugin_starter")
var ErrPluginNotFound = errors.New("invalid plugin_name")

func apiExec(req *common.ApiRequestData, ctx context.Context) error {
	var Package *data.Package
	plugin := data.GetPlugin(req.Package)
	if plugin != nil {
		Package = plugin.Package
	} else {
		for _, v := range data.ExternalPackages {
			if v.Name == req.Package {
				Package = v
			}
		}
	}
	if Package == nil {
		return ErrPackageNotFound
	}
	plugin = data.GetPlugin(req.Plugin)
	if plugin == nil {
		return ErrPluginNotFound
	}
	cmd, err := cmd.ToCmd(Package, plugin, req.Args[0], req.Args[1:], ctx)
	if err != nil {
		return err
	}
	return cmd.Run()
}
