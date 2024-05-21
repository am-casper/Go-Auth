# Go-Auth
Your one-stop guide to all Good Authentication methods via powers of GoLang and Gin framework. This branch contains JWT (JSON Web Token) + Refresh Token based Authentication. See <a href="https://github.com/am-casper/Go-Auth/tree/jwt#api-contents">API Contents</a> to understand the flow better.

# Setup
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
  &nbsp;&nbsp;&nbsp;&nbsp;"message" : "Login Successful" <br>
} <br>
      access-token and refresh-token in <b>RESPONSE HEADER</b>
    </td>
    <td>This endpoint logins the user in and return Access Token and Refresh Token in Cookies.</td>
  </tr>
  <tr>
  <td>/userInfo</td>
  <td>GET</td>
  <td> No Request Body <br> Add "Cookie": "access-token={access token}" to Request Header</td>
<td>
      {<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"username" : string, <br>
  &nbsp;&nbsp;&nbsp;&nbsp;"newsPref" : string,<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"moviePref" : string,<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"password" : string<br>
}  
    </td>
<td>This retrieves user information by using the Access Token for authentication.</td>
  </tr>
  <tr>
  <td>/refresh</td>
  <td>POST</td>
  <td> No Request Body <br> Add "Cookie": "refresh-token={access token}" to Request Header</td>
<td>
      {<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"message" : "Token Refreshed" <br>
}
    </td>
<td>This retrieves new Access Token and Refresh Token which shall be used for any API requests further. The older refresh tokens are now expired and can't be used anymore.</td>
  </tr>
  
</table>

The Access Token has a validity of 2 hours, after which the token expires. The client needs to refresh the access token using the refresh token so that the user can again have the access to his data. The validity of the refresh token in 24 hours. 

Have Fun Learning :)
