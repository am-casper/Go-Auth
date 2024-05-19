# Go-Auth
Your one-stop guide to all Good Authentication methods via powers of GoLang and Gin framework. This branch is the **base** branch, which consists the APIs without any Authentication techniques. This Repo does **not** contain any **main** branch. To see the code for technqiues, kidnly head over to the respective branched.

# Setup
The setup is really simple, you need to have a Mongo-DB account. Make a Database and within that, a collection. Make a copy of the .env.sample file using the following command in your terminal
```bash
$ > cp .env.sample .env
```
Enter the values of the environment variables. Do not doubt quote the strings. Start a connection from your MongoDB account and that's it! Just Run the following command to run the API over the port you've mentioned in the `.env` file.
```bash
$ > go run main.go
```


Have Fun Learning :)
