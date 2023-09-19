import { Card, CardContent, Typography } from "@mui/material"
const stripType = (type, address) => address.replace(`${type}.`, "")

const DeleteDetailCard = ({address, type, change, name}) => (<Card 
    sx={{minWidth: 320}}
    >
    <CardContent>
        <Typography sx={{fontSize: 14}} color="text.secondary" gutterBottom>
            {type}
        </Typography>
        <Typography variant="h5" component="div" color="warning.main">
        {/* {stripType(type, address)} */}
        {/* TODO: API /diffs/{changeSetId} needs to add before data */}
        {change.before ? change.before.name : `ref-{name}`}

        </Typography>
        <Typography sx={{ mb: 1.5}} color="text.secondary">
            {address}
        </Typography>
    </CardContent>
</Card>)

export default DeleteDetailCard