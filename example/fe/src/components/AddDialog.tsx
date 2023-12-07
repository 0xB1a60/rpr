import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import {useCallback, useState} from "react";
import {Alert, CircularProgress} from "@mui/material";
import ErrorIcon from '@mui/icons-material/Error';
import {HTTP_URL} from "../rpr/RPRConst.ts";

export const AddDialog = ({closeFunc}: { closeFunc: () => void }) => {
    const [value, setValue] = useState("");

    const [loading, setLoading] = useState(false);
    const [fetchError, setFetchError] = useState(false);
    const actionFunc = useCallback(async () => {
        setLoading(true);
        try {
            const response = await fetch(`${HTTP_URL}/add`, {
                method: "POST",
                body: JSON.stringify({
                    value: value,
                }),
            });
            if (response?.status !== 200) {
                throw Error(await response.text());
            }
        } catch (e) {
            console.log(e);
            setLoading(false);
            setFetchError(true);
            return;
        }

        closeFunc()
    }, [closeFunc, value]);

    return <Dialog open onClose={closeFunc} fullWidth>
        <DialogTitle>Add dialog</DialogTitle>
        <DialogContent dividers>
            {loading && <Alert icon={<CircularProgress/>} severity="info">
                Adding...
            </Alert>}

            {fetchError && <Alert icon={<ErrorIcon/>} severity="error">
                Error while adding, please check console
            </Alert>}

            <TextField
                autoFocus
                margin="dense"
                label="Value"
                type="text"
                fullWidth
                disabled={loading}
                value={value}
                onChange={(event) => setValue(event.target.value)}
                variant="standard"
            />
        </DialogContent>
        <DialogActions>
            <Button onClick={closeFunc}>Cancel</Button>
            <Button onClick={actionFunc} disabled={value.length === 0 || loading || value.length > 10_000}
                    variant={"contained"}>Add</Button>
        </DialogActions>
    </Dialog>;
};
