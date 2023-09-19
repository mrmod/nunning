import { Card, CardContent, Typography } from "@mui/material"
import ChangeDiffs from "./ChangeDiffs";
const stripType = (type, address) => address.replace(`${type}.`, "")

const UpdateDetailCard = ({address, type, name, mode, change, changeDiffs}) => (<Card sx={{minWidth: 320}}>
    <CardContent>
        <Typography sx={{fontSize: 14}} color="text.secondary" gutterBottom>
            {type}
        </Typography>
        <Typography variant="h5" component="div">
            {/* {stripType(type, address)} */}
            {change.before ? change.before.name : `ref-{name}`}
        </Typography>
        <Typography sx={{ mb: 1.5}} color="text.secondary">
            {name}
        </Typography>
        <ChangeDiffs changeDiffs={changeDiffs} />
    </CardContent>
</Card>)

export default UpdateDetailCard