import { Typography } from "@mui/material";

const ChangeSet = ({changeSetId}) => (<>
    <Typography variant={"h5"}>ChangeSet</Typography>
    <Typography flex-align={"right"} variant={"body2"}>{changeSetId}</Typography>
</>)
export default ChangeSet