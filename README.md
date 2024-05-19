# Go-Auth
Your one-stop guide to all Good Authentication methods via powers of GoLang and Gin framework. This branch is based on the **Bearer Token Based Authentication**. At the frontend, we have to append the access token to the headers for the API calls to authenticate and authorize the user for his contents.

## Setup
The setup is really simple, you need to have a Mongo-DB account. Make a Database and within that, a collection. Make a copy of the .env.sample file using the following command in your terminal
```bash
$ > cp .env.sample .env
```
Enter the values of the environment variables. Do not doubt quote the strings. Start a connection from your MongoDB account and that's it! Just Run the following command to run the API over the port you've mentioned in the `.env` file.
```bash
$ > go run main.go
```

## API Contents

<table>
  <th>ENDPOINTS</th>
  <th>HTTP REQUEST METHOD</th>
  <th>REQUEST BODY</th>
  <th>RESPONSE BODY</th>
  <th>DESCRIPTION</th>
  <tr>
    <td>/register</td>
    <td>POST</td>
    <td>
      {<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"username" : string, <br>
  &nbsp;&nbsp;&nbsp;&nbsp;"newsPref" : string,<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"moviePref" : string,<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"password" : string<br>
}  
    </td>
    <td> A JSON object <br> with hashed password <br> confirms Success.</td>
    <td> This endpoint registers the user.</td>
  </tr>
  <tr>
    <td>/login</td>
    <td>POST</td>
    <td>
      {<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"username" : string, <br>
  &nbsp;&nbsp;&nbsp;&nbsp;"password" : string <br>
}  
    </td>
    <td>
     {<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"accessToken" : string <br>
}  
    </td>
    <td>This endpoint logins the user in and return an Access Token that the frontend needs to append to the headers of future API calls to access the data authentically.</td>
  </tr>
  <tr>
  <td>/userInfo</td>
  <td>GET</td>
  <td> No Request Body <br> Add "Authorization": "Bearer {access token}>" to Request Header</td>
<td>
      {<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"username" : string, <br>
  &nbsp;&nbsp;&nbsp;&nbsp;"newsPref" : string,<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"moviePref" : string,<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"password" : string<br>
}  
    </td>
<td>This retrieves user information by using the Bearer Access Token for authentication.</td>
  </tr>

  
</table>

That wraps up a Basic API using Bearer Access Token Technique.

Have Fun Learning :)
