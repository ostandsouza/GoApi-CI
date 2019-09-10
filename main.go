package main

import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"
    "io/ioutil"
    "path"

    mongo "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/gridfs"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/gorilla/mux"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/x/bsonx"
)

var client *mongo.Client
var db *mongo.Database
var bucket *gridfs.Bucket
var ustream *gridfs.UploadStream
var dstream *gridfs.DownloadStream
var fileName string

//var wg sync.WaitGroup

//React is structure in which data is stored at database.
type React struct {
    ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Time string `json:"time,omitempty" bson:"time,omitempty"`
    SuiteName string `json:"suitename,omitempty" bson:"suitename,omitempty"`
    Device string `json:"device,omitempty" bson:"device,omitempty"`
    Notes string `json:"notes,omitempty" bson:"notes,omitempty"`
    FailedTestCases string `json:"FailedTestCases,omitempty" bson:"FailedTestCases,omitempty"`
    PassedTestCases string `json:"PassedTestCases,omitempty" bson:"PassedTestCases,omitempty"`
    SkippedTestCases string `json:"SkippedTestCases,omitempty" bson:"SkippedTestCases,omitempty"`
    TotalTestCases string `json:"TotalTestCases,omitempty" bson:"TotalTestCases,omitempty"`
    RunType string `json:"RunType,omitempty" bson:"RunType,omitempty"`
    UIreport string `json:"uiReport,omitempty" bson:"uiReport,omitempty"`
    FunctionalReport string `json:"functionalReport,omitempty" bson:"functionalReport,omitempty"`
    Version string `json:"version,omitempty" bson:"version,omitempty"`
}

//MongoImage is structure in which files is stored at database.
type MongoImage struct {
    ID primitive.ObjectID `bson:"_id"`
    UploadDate primitive.DateTime `bson:"uploadDate"`
    FileSize int64 `bson:"chunkSize"`
    Length int `bson:"length"`
    Filename string `json:"filename,omitempty" bson:"filename"`
    metadata bsonx.Doc `bson:"metadata"`
}

func handler(response *http.ResponseWriter) {
    //defer wg.Done()
    p := *response
    if db == nil {
        fmt.Println("handler")
        v := make(map[string]string)
        v["status"] = "Connection to the server was not successful"

        json.NewEncoder(p).Encode(v)
        //p.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        // if r := recover(); r != nil {
        //  fmt.Println("Recovered in f", r)
        //  p.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        //  if _, ok := r.(runtime.Error); ok {
        //      panic(r)
        //  }
        // }
        //  panic("server not connected")
    }
}

//GetAllCollections returns the name of all collection in the given database
func GetAllCollections(response http.ResponseWriter, request *http.Request) {

    response.Header().Set("content-type", "application/json")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }

    react, err := db.ListCollectionNames(ctx, bson.D{})
    fmt.Println("cur", react)
    if err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }

    // response.Header().Set("content-type", "application/json")
    // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    // defer cancel()
    // //react, err := client.Database("react").ListCollectionNames(ctx, bson.M{})
    // //wg.Add(1)
    // //defer handler(&response)
    // react, err := db.ListCollectionNames(ctx, bson.D{})
    // fmt.Println("err", err.Error())
    // if err != nil {
    //  response.WriteHeader(http.StatusInternalServerError)
    //  response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
    //  return
    // }
    v := make(map[string][]string)
    v["status"] = react

    json.NewEncoder(response).Encode(v)
}

//GetAllDocuments returns the all documents from collection in the given database
func GetAllDocuments(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")

    type PostBody struct {
        API string `json:"api,omitempty"`
    }

    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    //params := mux.Vars(request)
    //api := params["api"]
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    cursor, errors := collection.Find(ctx, bson.M{})
    if errors != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "` + errors.Error() + `" }`))
        return
    }
    if api == "IssueMap" {
        var react []interface{}
        var data interface{}
        v := make(map[string][]interface{})
        defer cursor.Close(ctx)
        for cursor.Next(ctx) {
            cursor.Decode(&data)
            react = append(react, data)
        }
        v["status"] = react
        json.NewEncoder(response).Encode(v)

    } else {
        var react []React
        var data React
        v := make(map[string][]React)
        defer cursor.Close(ctx)
        for cursor.Next(ctx) {
            cursor.Decode(&data)
            react = append(react, data)
        }
        v["status"] = react
        json.NewEncoder(response).Encode(v)

    }
}

//GetConnStatus is used to connect to mongo service
func GetConnStatus(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")

    type PostBody struct {
        URL string `json:"url,omitempty"`
    }
    v := make(map[string]string)

    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    url := t.URL
    fmt.Println("url", url)
    re := regexp.MustCompile(`^([a-zA-Z]+)://([a-zA-Z\d\.]+):([\d]+)/`)
    uri := re.FindString(url)
    dbName := re.ReplaceAllString(url, "")
    fmt.Println("uri", uri)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client, errors := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if errors != nil {
        fmt.Println("dbName", dbName)
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Not connected to server" }`))
        return
    }

    db := client.Database(dbName)
    fmt.Println("new", db)
    setRef(db)

    react, _ := db.ListCollectionNames(ctx, bson.D{})
    fmt.Println("cur", react)
    for _, n := range react {
        if "IssueMap" == n {
            response.Write([]byte(`{ "status": "connected to server" }`))
            return
        }
    }
    response.WriteHeader(http.StatusInternalServerError)
    v["status"] = "Not connected to ReactDB"

    json.NewEncoder(response).Encode(v)
}

//GetConnStatusFlag indicates whether the database connection is successful
func GetConnStatusFlag(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    react, err := db.ListCollectionNames(ctx, bson.D{})
    fmt.Println("cur", react)
    if err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "false" }`))
        return
    }
    response.Write([]byte(`{ "status": "true" }`))
    return
}

//CreateCollection is used to add new collection to database
func CreateCollection(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    type PostBody struct {
        API string `json:"api,omitempty"`
    }

    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    fmt.Println("Name", api)
    //params := mux.Vars(request)
    //api := params["api"]
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    _, errors := collection.InsertOne(ctx, bson.D{})
    //fmt.println(res.InsertedID.(primitive.ObjectID).Hex())
    if errors != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "` + errors.Error() + `" }`))
        return
    }

    v := make(map[string]string)
    v["status"] = "Collection creation was successfull"

    json.NewEncoder(response).Encode(v)
}

//DeleteCollection is used to delete existing collection to database
func DeleteCollection(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    type PostBody struct {
        API string `json:"api,omitempty"`
    }

    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    fmt.Println("Name", api)
    //params := mux.Vars(request)
    //api := params["api"]
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    errors := collection.Drop(ctx)
    if errors != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "` + errors.Error() + `" }`))
        return
    }

    v := make(map[string]string)
    v["status"] = "Collection was deleted"

    json.NewEncoder(response).Encode(v)
}

//GetOneDocument returns the one documents from collection depending on filter given
func GetOneDocument(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")

    type PostBody struct {
        API string `json:"api,omitempty"`
        ID string `json:"id,omitempty"`
    }
    var react React
    //res.InsertedID.(primitive.ObjectID).Hex()
    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    id, _ := primitive.ObjectIDFromHex(t.ID)
    filter := bson.M{"_id": id}
    fmt.Println("filter", filter)
    //params := mux.Vars(request)
    //api := params["api"]
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    res := collection.FindOne(ctx, filter)
    res.Decode(&react)
    fmt.Println(react)
    if res.Err() != nil {
        fmt.Println("error", res.Err())
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to find document" }`))
        return
    }
    v := make(map[string]React)
    v["status"] = react

    json.NewEncoder(response).Encode(v)
}

//InsertDocument is used to enter one documents to collection
func InsertDocument(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    type PostBody struct {
        API string `json:"api,omitempty"`
        Doc map[string]string `json:"insertDoc,omitempty"`
    }
    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    fmt.Print("t",t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "bstatus": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    doc := t.Doc
    // doc, errorj := json.Marshal(data)
    // if errorj != nil {
    //  response.WriteHeader(http.StatusInternalServerError)
    //  fmt.Println("errorj", errorj.Error())
    // }
    fmt.Println("Name", api)
    fmt.Println("doc", doc)
    //params := mux.Vars(request)
    //api := params["api"]
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    res, errors := collection.InsertOne(ctx, doc)
    if errors != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "cstatus": "` + errors.Error() + `" }`))
        return
    }
    //fmt.println(res.InsertedID.(primitive.ObjectID).Hex())
    fmt.Println(res.InsertedID)
    fmt.Println(res.InsertedID.(primitive.ObjectID).Hex())

    v := make(map[string]string)
    v["status"] = "document were successfully inserted at Id:" + res.InsertedID.(primitive.ObjectID).Hex()

    json.NewEncoder(response).Encode(v)
}

//UpdateDocument updates one documents from collection depending on filter given
func UpdateDocument(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    type PostBody struct {
        API string `json:"api,omitempty"`
        ID string `json:"id,omitempty"`
        Replace map[string]string `json:"replace,omitempty"`
    }

    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    id, _ := primitive.ObjectIDFromHex(t.ID)
    filter := bson.M{"_id": id}
    replace := t.Replace
    fmt.Println("Name", api)
    fmt.Println("filter", filter)
    update := bson.M{"$set": replace}

    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    res, errors := collection.UpdateOne(ctx, filter, update)
    if errors != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "` + errors.Error() + `" }`))
        return
    }
    fmt.Println("MatchedCount", res.MatchedCount)
    fmt.Println("ModifiedCount", res.ModifiedCount)

    v := make(map[string]string)
    v["status"] = "No of documents modified successfully:" + strconv.FormatInt(res.ModifiedCount, 10)

    json.NewEncoder(response).Encode(v)
}

//DeleteEntry delete one entry from document depending on filter given
func DeleteEntry(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    type PostBody struct {
        API string `json:"api,omitempty"`
        ID string `json:"id,omitempty"`
        Replace map[string]string `json:"replace,omitempty"`
    }

    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    id, _ := primitive.ObjectIDFromHex(t.ID)
    filter := bson.M{"_id": id}
    replace := t.Replace
    fmt.Println("Name", api)
    fmt.Println("filter", filter)
    update := bson.M{"$unset": replace}

    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    res, errors := collection.UpdateOne(ctx, filter, update)
    if errors != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "` + errors.Error() + `" }`))
        return
    }
    fmt.Println("MatchedCount", res.MatchedCount)
    fmt.Println("ModifiedCount", res.ModifiedCount)

    v := make(map[string]string)
    v["status"] = "No of documents modified successfully:" + strconv.FormatInt(res.ModifiedCount, 10)

    json.NewEncoder(response).Encode(v)
}

//DeleteDocument is used to delete one document from collection depending on filter given
func DeleteDocument(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")

    type PostBody struct {
        API string `json:"api,omitempty"`
        ID string `json:"id,omitempty"`
    }
    //res.InsertedID.(primitive.ObjectID).Hex()
    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    id, _ := primitive.ObjectIDFromHex(t.ID)
    filter := bson.M{"_id": id}
    fmt.Println("filter", filter)
    //params := mux.Vars(request)
    //api := params["api"]
    if db == nil {
        response.Write([]byte(`{ "status": "Connection to the server was not successful" }`))
        return
    }
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    res, errors := collection.DeleteOne(ctx, filter)
    if errors != nil {
        fmt.Println("error", errors.Error())
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to find document" }`))
        return
    }
    fmt.Println(res.DeletedCount)

    v := make(map[string]string)
    v["status"] = "No of documents deleted successfully:" + strconv.FormatInt(res.DeletedCount, 10)
    json.NewEncoder(response).Encode(v)
}

//CountDocument is used to get count of documents from collection depending on filter given
func CountDocument(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    type PostBody struct {
        API string `json:"api,omitempty"`
    }
    //res.InsertedID.(primitive.ObjectID).Hex()
    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    api := t.API
    collection := db.Collection(api)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    res, errors := collection.CountDocuments(ctx, bson.D{})
    if errors != nil {
        fmt.Println("error", errors.Error())
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to count document" }`))
        return
    }
    fmt.Println(res)
    v := make(map[string]string)
    v["status"] = "No of documents in the collection:" + strconv.FormatInt(res, 10)
    json.NewEncoder(response).Encode(v)
}

//UploadDocument is used to upload documents to the server
func UploadDocument(response http.ResponseWriter, request *http.Request) {
    //response.Header().Set("content-type", "application/json")
    file, handler, err := request.FormFile("file")
    if err != nil {
        fmt.Println("err", err)
        http.Error(response, "Error uploading file", http.StatusInternalServerError)
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to upload document" }`))
        return
    }
    defer file.Close()
    name := strings.Split(handler.Filename, ".")
    fmt.Printf("File name %s\n", name[0])
    f, errors := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
    if errors != nil {
        fmt.Println("err", errors)
        http.Error(response, "Error coopying file", http.StatusInternalServerError)
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to copy document" }`))
        return
    }
    defer f.Close()
    io.Copy(f, file)
    v := make(map[string]string)
    v["status"] = "file successfully uploaded"
    json.NewEncoder(response).Encode(v)
}

//MultiUploadDocument is used to upload documents to the server
func MultiUploadDocument(response http.ResponseWriter, request *http.Request) {
    request.ParseMultipartForm(32 << 20) // 32MB is the default used by FormFile
    fhs := request.MultipartForm.File["file"]
    for _, fh := range fhs {
        file, err := fh.Open()
        if err != nil {
            fmt.Println("err", err)
            http.Error(response, "Error uploading file", http.StatusInternalServerError)
            response.WriteHeader(http.StatusInternalServerError)
            response.Write([]byte(`{ "status": "Unable to upload document" }`))
            return
        }
        defer file.Close()
        dt := time.Now()
        newpath := filepath.Join(".", "results/"+dt.Format("01-02-2006"))
        os.MkdirAll(newpath, os.ModePerm)
        name := strings.Split(fh.Filename, ".")
        fmt.Printf("File name %s\n", name[0])
        f, errors := os.OpenFile(newpath+"/"+fh.Filename, os.O_WRONLY|os.O_CREATE, 0666)
        if errors != nil {
            fmt.Println("err", errors)
            http.Error(response, "Error coopying file", http.StatusInternalServerError)
            response.WriteHeader(http.StatusInternalServerError)
            response.Write([]byte(`{ "status": "Unable to copy document" }`))
            return
        }
        defer f.Close()
        io.Copy(f, file)
    }
}

//UploadDocuments is used to upload documents to the server
func UploadDocuments(response http.ResponseWriter, request *http.Request) {
    file, handler, err := request.FormFile("file")
    if err != nil {
        fmt.Println("err", err)
        http.Error(response, "Error uploading file", http.StatusInternalServerError)
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to upload document" }`))
        return
    }
    defer file.Close()
    name := strings.Split(handler.Filename, ".")
    fmt.Printf("File name %s\n", name[0])

    if bucket, err = gridfs.NewBucket(db, options.GridFSBucket().SetName("file")); err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to set collection name" }`))
        return
    }
    opts := options.GridFSUpload()
    opts.SetMetadata(bsonx.Doc{{Key: "content-type", Value: bsonx.String(handler.Header.Get("content-type"))}})
    if ustream, err = bucket.OpenUploadStream(handler.Filename, opts); err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to open file" }`))
        return
    }
    filem, err := handler.Open()

    r := bufio.NewReader(filem)
    buf := make([]byte, 1024)
    for {
        n, err := r.Read(buf)
        if err != nil && err != io.EOF {
            response.WriteHeader(http.StatusInternalServerError)
            response.Write([]byte(`{ "status": "Unable to read document to bufio" }`))
            return
        }
        if n == 0 {
            break
        }
        if _, err := ustream.Write(buf[:n]); err != nil {
            response.WriteHeader(http.StatusInternalServerError)
            response.Write([]byte(`{ "status": "Unable to write document to bufio" }`))
            return
        }
    }

    fileID := ustream.FileID
    fmt.Println("fileID", fileID)
    ustream.Close()
    var bt bytes.Buffer
    w := bufio.NewWriter(&bt)
    res, err := bucket.DownloadToStream(fileID, w)
    if err != nil {
        fmt.Println(err)
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to upload file" }`))
        return
    }
    v := make(map[string]interface{})
    fmt.Println("res", res)
    v["status"] = fileID
    json.NewEncoder(response).Encode(fileID)
}

//DownloadDocuments is used to download documents from the server
func DownloadDocuments(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")

    type PostBody struct {
        ID string `json:"id,omitempty"`
    }
    //res.InsertedID.(primitive.ObjectID).Hex()
    decoder := json.NewDecoder(request.Body)
    var t PostBody
    err := decoder.Decode(&t)
    if err != nil {
        //panic(err)
        response.Write([]byte(`{ "status": "` + err.Error() + `" }`))
        return
    }
    id, _ := primitive.ObjectIDFromHex(t.ID)
    filter := bson.M{"_id": id}

    if bucket, err = gridfs.NewBucket(db, options.GridFSBucket().SetName("file")); err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to set collection nam" }`))
        return
    }
    opts := options.GridFSFind()
    //opts.SetLimit
    var react []MongoImage
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    cursor, errors := bucket.Find(filter, opts)
    if errors != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "` + errors.Error() + `" }`))
        return
    }

    defer cursor.Close(ctx)
    for cursor.Next(ctx) {
        //var data React
        var data MongoImage
        cursor.Decode(&data)
        react = append(react, data)
    }
    fmt.Println("react", react)

    if dstream, err = bucket.OpenDownloadStream(react[0].ID); err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        response.Write([]byte(`{ "status": "Unable to read file" }`))
        return
    }
    fmt.Println("fileID", react[0].Filename)

    newpath := filepath.Join(".", "downloads/")
    //os.RemoveAll("/downloads")
    dir, direrr := ioutil.ReadDir("./downloads")
    for _, d := range dir {
        os.RemoveAll(path.Join([]string{"downloads", d.Name()}...))
    }
    os.MkdirAll(newpath, os.ModePerm)

    if direrr != nil {
        fmt.Println("not able delete downloads folder")
    }

    f, err := os.OpenFile(newpath+"/"+react[0].Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        response.Write([]byte(`{ "status": "Unable create a file" }`))
        return
    }

    r := bufio.NewReader(dstream)
    buf := make([]byte, 1024)
    for {
        n, err := r.Read(buf)
        if err != nil && err != io.EOF {
            response.WriteHeader(http.StatusInternalServerError)
            response.Write([]byte(`{ "status": "Unable to read document to bufio" }`))
            return
        }
        if n == 0 {
            break
        }
        if _, err := f.Write(buf[:n]); err != nil {
            response.WriteHeader(http.StatusInternalServerError)
            response.Write([]byte(`{ "status": "Unable to write document to bufio" }`))
            return
        }
    }
    fileName = react[0].Filename
    fmt.Println("fileName",fileName)
    v := make(map[string]string)
    v["status"] = react[0].Filename
    json.NewEncoder(response).Encode(v)
}

func setRef(db1 *mongo.Database) {
    db = db1
}

func getRef() *mongo.Database {
    return db
}

func main() {
    fmt.Println("Starting the application...")
    router := mux.NewRouter()
    router.HandleFunc("/api/list", GetAllCollections).Methods("GET")
    router.HandleFunc("/api/read", GetAllDocuments).Methods("POST")
    router.HandleFunc("/api/", GetConnStatus).Methods("POST")
    router.HandleFunc("/api/conn", GetConnStatusFlag).Methods("GET")
    router.HandleFunc("/api/new", CreateCollection).Methods("POST")
    router.HandleFunc("/api/drop", DeleteCollection).Methods("POST")
    router.HandleFunc("/api/find", GetOneDocument).Methods("POST")
    router.HandleFunc("/api/create", InsertDocument).Methods("POST")
    router.HandleFunc("/api/update", UpdateDocument).Methods("POST")
    router.HandleFunc("/api/unset", DeleteEntry).Methods("POST")
    router.HandleFunc("/api/delete", DeleteDocument).Methods("POST")
    router.HandleFunc("/api/count", CountDocument).Methods("POST")
    router.HandleFunc("/api/upload1", UploadDocument).Methods("POST")
    router.HandleFunc("/api/multiupload", MultiUploadDocument).Methods("POST")
    router.HandleFunc("/api/upload", UploadDocuments).Methods("POST")
    router.HandleFunc("/api/download", DownloadDocuments).Methods("POST")
    router.HandleFunc("/file/", func(response http.ResponseWriter, request *http.Request) {
        http.ServeFile(response, request, "./downloads/"+fileName)
    }).Methods("GET")
    // http.Handle("/static/", http.StripPrefix(strings.TrimRight("/statics/", "/"), http.FileServer(http.Dir("./build"))))
    http.Handle("/static/", http.FileServer(http.Dir("./build")))
    log.Fatal(http.ListenAndServe(":4100", router))
}