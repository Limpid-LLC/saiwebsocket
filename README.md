# saiWebSocket
WebSockets message broker
# Start SaiWebSocket Server
 $ ./websockets_pro 
# Config
You should have an configuration file websockets.conf in the current working directory. Here is a full example of configuration file:
```
{
  Host: "localhost",
  Port: "8000",
  Origin: "*",
  Responseheaders: "*",
# comment this string if you don't want to allow connection for a clients without preregistered tokens
# in case you leave this line each client connected without a token will obtain token "unregistered"
# until the client send sistem message "RegisterToken:<SomeToken>"
  AllowUnregisteredClients: "yes", 
# comment this line if you don't want to download registered tokens
# here is the example of url output
# {"abc1":true,"abc2":true,"abc3":true,"group\/item1":true,"group\/item2":true}
  RegisteredTokensUrl: "http://saiset.co/registeredtokens.php", 
} 
```

# Connection From Web Client (Browser) With Client JS
Connection with authorisation
``` 
	var socket = new WebSocket("ws://localhost:8000/ws?RegisterToken=abc1");
```          

Example1
``` 
	var input = document.getElementById("input");
	var output = document.getElementById("output");
	var socket = new WebSocket("ws://localhost:8000/ws?RegisterToken=abc1");

	socket.onopen = function () {
		output.innerHTML += "Status: Connected\n";
	};

	socket.onmessage = function (e) {
		output.innerHTML += "Server: " + e.data + "\n";
	};

	function send() {
		socket.send(input.value);
		input.value = "";
	}
```          

Connection without authorisation
``` 
	var socket = new WebSocket("ws://localhost:8000/ws");
```          

Example2
``` 
	var input = document.getElementById("input");
	var output = document.getElementById("output");
	var socket = new WebSocket("ws://localhost:8000/ws");

	socket.onopen = function () {
		output.innerHTML += "Status: Connected\n";
		socket.send("RegisterToken:abc1");
	};

	socket.onmessage = function (e) {
		output.innerHTML += "Server: " + e.data + "\n";
	};

	function send() {
		socket.send(input.value);
		input.value = "";
	}
```
          
# Message Structure
The message can consists of two parts beaked with the delimiter symbol "|". The first part is a list of receivers tokens. The second part is a message itself. In case the message haven't delimiter "|" the message typify as a service message.
Message example:
```
 abc1,group/,abc3|Hello 
```
"Hello" message will delivered to the clients with the following tokens from the config file example:\
abc1\
abc3\
group/item1\
group/item2

# Service Messages
Register token from webclient side:
``` RegisterToken:token ```

Echo service:
``` Echo:message ```

# API
- Broadcast Message
  - Parameters
    * method : broadcast
    * message : message
- Get Connected Clients List
  - Parameters
    * method : get_clients
    * message : message
- Register Token
  - Parameters
    * method : registerToken
    * message : token
- Unregister Token
  - Parameters
    * method : unregisterToken
    * message : troken
- Get Registered Token List
  - Parameters
    * method : registeredTokenList
