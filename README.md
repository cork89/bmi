# National Dishes and BMI Gallery

A simple website to display the information from the following wikipedia articles:
- https://en.wikipedia.org/wiki/List_of_countries_by_body_mass_index
- https://en.wikipedia.org/wiki/National_dish


## Tech Stack

- **Backend**: Go
- **Frontend**: Standard Library `html/template`, JavaScript
- **Data**: CSV-based storage
- **Environment**: `godotenv` for configuration

## Prerequisites

- Go 1.16 or higher
- A `.env` file in the root directory
- A `countries.csv` file in the root directory

## Setup

1.  **Environment Configuration**  
    Create a `.env` file in the root directory and define your image source (CDN or local path):
    ```text
    IMG_SOURCE=https://your-image-bucket.com
    ```

2.  **Data Format**  
    Ensure your `countries.csv` follows this column structure:
    - 0: Country Name
    - 1: BMI Value (Float)
    - 4: National Dish Name
    - 5: Wikipedia Link
    - 6: Image Filename
    - 7: Image Aspect Ratio (Float)
    - 8: Order (2 Columns)
    - 9: Order (3 Columns)
    - 10: Order (4 Columns)

3.  **Run the Application**
    ```bash
    go run main.go
    ```
    or
    ```bash
    air
    ```
    The server will start on `http://localhost:8083`.

## Project Structure

- `main.go`: Handles CSV parsing, template rendering, and API logic for content updates.
- `static/`:
    - `base.html`: The core HTML boilerplate.
    - `home.html`: The main view containing the JavaScript observers and UI components.
    - `content.html`: The partial template used for rendering the gallery grid.
    - `style.css`: styling
