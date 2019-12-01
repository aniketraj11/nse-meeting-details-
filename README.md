# nse-meeting-details-
Standalone go app to get the details about the board meetings from the NSE website

- This project is developed in Go. 

- The user is provided with a search box and a button. 

- The search box accepts stock's trading symbol as well as the company name and this sends AJAX request to the backend for getting the company name and symbol from the https://www.nseindia.com/ website.

- On sending the request, the backend scrapes the NSE page and displays the following results: Board meeting date, Purpose, Details 
