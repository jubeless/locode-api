# Locode API

This is a simple API that serves location data based on UN/LOCODE.

## Prerequisites

- Go (version 1.13 or later recommended)

## Running the API

1.**Download the CSV Data:**

You will need to obtain the `locode_1.csv`, `locode_2.csv`, and `locode_3.csv` files. These files should contain the UN/LOCODE data in CSV format. Place these files in the same directory as your `main.go` file. The structure of the csv files is important. The program expects the following structure:

- The program identifies country entries when a row has the format: `"", "CountryCode", "", "CountryName", ...`.  It extracts `CountryCode` and `CountryName` from such rows.
- Location entries should have at least 11 columns.
- The relevant columns for location entries are:
    - Column 2: `AdminCode`
    - Column 3: `LocationCode`
    - Column 4: `Name`
    - Column 5: `AltName`
    - Column 11: `Coordinates`
- The `CountryCode` is taken from the last encountered country entry row, or from column 2 if no country row has been encountered.

2.**Navigate to the Directory:**

Open a terminal and navigate to the directory where you saved `main.go` and the CSV files.

```bash
cd path/to/your/directory
```

3.**Run the API:**

Use the `go run` command to compile and run the `main.go` file.

```bash
go run main.go
```

## API Endpoints

The API has a single endpoint:

- `/locode?locode={locode}`

    This endpoint retrieves location information for a given UN/LOCODE.  The `{locode}` parameter should be replaced with the actual LOCODE you want to query.  The LOCODE is a combination of the `CountryCode` and the `LocationCode`.

- `/random?count=10`

    This endpoint returns a JSON array of <count> random locations from the dataset. If count is not provided, it defaults to 3000.

## Using curl to Interact with the API

Here are a few examples of how to use `curl` to interact with the API:

1.**Retrieve location information for a specific LOCODE:**

```bash
curl "http://localhost:5555/locode?locode=USNYC"
```

Replace `USNYC` with a valid locode from your dataset. This will return a JSON response containing the location information, for example:

```json
{
  "locode": "USNYC",
  "country_code": "US",
  "country_name": "united states of america (the)",
  "admin_code": "US",
  "location_code": "NYC",
  "name": "New York",
  "alt_name": "New York",
  "coordinates": "4042N 07400W"
}
```

2.**Query with an unknown LOCODE:**

```bash
curl "http://localhost:5555/locode?locode=INVALID"
```

This will return a `404 Not Found` error:

```txt
Unknown locode INVALID
```

3.**Query without a LOCODE parameter:**

```bash
curl "http://localhost:5555/locode"
```

This will return a `400 Bad Request` error:

```txt
Missing 'locode' parameter
```
