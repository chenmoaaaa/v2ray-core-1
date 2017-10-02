package all

import (
	// The following are necessary as they register handlers in their init functions.
	_ "github.com/whatedcgveg/v2ray-core/app/dispatcher/impl"
	_ "github.com/whatedcgveg/v2ray-core/app/dns/server"
	_ "github.com/whatedcgveg/v2ray-core/app/proxyman/inbound"
	_ "github.com/whatedcgveg/v2ray-core/app/proxyman/outbound"
	_ "github.com/whatedcgveg/v2ray-core/app/router"

	_ "github.com/whatedcgveg/v2ray-core/proxy/blackhole"
	_ "github.com/whatedcgveg/v2ray-core/proxy/dokodemo"
	_ "github.com/whatedcgveg/v2ray-core/proxy/freedom"
	_ "github.com/whatedcgveg/v2ray-core/proxy/http"
	_ "github.com/whatedcgveg/v2ray-core/proxy/shadowsocks"
	_ "github.com/whatedcgveg/v2ray-core/proxy/socks"
	_ "github.com/whatedcgveg/v2ray-core/proxy/vmess/inbound"
	_ "github.com/whatedcgveg/v2ray-core/proxy/vmess/outbound"

	_ "github.com/whatedcgveg/v2ray-core/transport/internet/kcp"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/tcp"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/tls"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/udp"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/websocket"

	_ "github.com/whatedcgveg/v2ray-core/transport/internet/headers/http"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/headers/noop"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/headers/srtp"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/headers/utp"
	_ "github.com/whatedcgveg/v2ray-core/transport/internet/headers/wechat"
)
