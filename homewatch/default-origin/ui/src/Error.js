import { Box } from "@chakra-ui/react"
class ApiError {
    constructor(fetchResponse) {
        this.status = fetchResponse.status
        this.text = null
        fetchResponse.text().then(text => this.text = text) 
    }
    error() {
        try {
            return JSON.parse(this.text)
        } catch(err) {
            return this.text
        }
    }
}
const Error = ({error}) => {
    try {
        return (<Box>
            Error: {JSON.stringify(error)}
        </Box>)
    } catch(err) {
        console.log("Failed to create <Error /> from error: ", err)
        return (<Box>
            Unable to create error. Check console logs
        </Box>)
    }
}

export default Error
export {ApiError}