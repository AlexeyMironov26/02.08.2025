My server use http as a main protocol for interaction with it and the server is launching on 0.0.0.0 (means that my server after launching accepts and processes http requests on all ip-addresses which this machine possesses) (IP):8080(port) it is for shaping zip archives from files found on internet and providing a link for downoading that archive. the example of url to address to host: http://localhost(can be any other from ips of your pc):8080/
Launch it as:
cd stuff_load
go run main/mainload.go 
in one terminal and send http requests on api of server from another(you can utility "curl" for that). To shut down the server push ctrl+c into the terminal where it was ran.

Server have the following endpoints:

1). /ini_arch
you should send post request on this endpoint to initialize new task by creating an archive and as a response you'll obtain the id of this new task. Not more than 3 tasks can be in progress simultaneously (but nevertheless you able to get access to unlimited amount of ready archives by id of its tasks sending requests at endpoint /arch_stat)
example: curl -X POST "http://localhost:8080/ini_arch"

2). /add_obj
    this endpoint is for adding objects in particular archive, which we already began to form in our task. You should specify as parameter in url of your request the id of task, to which archive you wish to add your object (only .jpeg and .pdf formats are permited). One archive should contain 3 files. The body of patch request should have header Content-type : application/json and contain a url from the internet on resourse like {
        "url": "https://site.com/imageee3576888.jpeg"
    }
    example: curl -X PATCH -H "Content-Type: application/json" -d '{"url":"https://somesite.com/doc67.pdf"}' "http://localhost:8080/add_obj?id=3"
    cmd version: curl -X PATCH -H "Content-Type: application/json" -d "{\"url\":\"https://avatars.mds.yandex.net/i?id=c0ce12577ec83ca557f3ed199f9539d7_l-5162829-images-thumbs^&n=13\"}" "http://localhost:8080/add_obj?id=2"

3). /arch_stat
you need to send on this endpoint get request with parameter "id" in url of request, which will point on a particular task (remmeber the ids of all your tasks when you create it sending request on endpoint /ini_arch) and in response you will get the status of task: "initialized"-archive without any objects, "in progress" - archive have 1-2 objects - or "ready" - archive have 3 loaded objects, in this situation with status you get as well a link for downloading this archive.
example: curl "http://localhost:8080/arch_stat?id=3"
4) /archive/{id of ended task} - it is a group of endpoints including links on all formed archives, managed by one handler with endpoint /archive/. When after checking status of archive you find it "ready", you get the special archive link on this server, to get the archive-zip file in body of response you should send a get-request on that special link. 
example: You can download it on your computer with different ways, like with help of curl which will download the obtained data from get request on specified url: curl -L -O http://localhost:8080/archive/3. Or you can download the archive by just ctrl+click if your terminal lets you and awtomaticly will be sent get request on handler and will be obtained a response with whole resourse data which your pc automatically will propose to save as zip-archive due to special Headers 
