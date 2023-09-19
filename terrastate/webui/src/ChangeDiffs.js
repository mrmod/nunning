import { List, ListItem, ListItemText, Typography } from "@mui/material";

const ChangeDiffs = ({changeDiffs}) => (<List>
    {changeDiffs.map((diff, key) => (<ChangeDiff diff={diff} key={`diff-${key}-${diff.property}`} />))}
</List>)

const ChangeDiff = ({diff}) => (<ListItem>
    <ListItemText
        primary={diff.property}
        secondary={<>
            <Typography variant="body2" component="span" color={"error.main"}>{diff.from}</Typography>
        </>}
    />
        
    <ListItemText
        primary="now"
        secondary={<>
            <Typography variant="body2" component="span" color={"success.main"}>{diff.to}</Typography>
        </>}
    />
</ListItem>)

export default ChangeDiffs