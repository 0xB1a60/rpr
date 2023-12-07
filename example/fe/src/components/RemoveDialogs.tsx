import {useCallback, useState} from "react";
import {Alert, CircularProgress, Dialog, DialogContent} from "@mui/material";
import Button from "@mui/material/Button";
import DialogActions from "@mui/material/DialogActions";
import DialogTitle from "@mui/material/DialogTitle";
import ErrorIcon from "@mui/icons-material/Error";
import {HTTP_URL} from "../rpr/RPRConst.ts";

export const RemoveDialog = ({id, closeFunc}: { id: string, closeFunc: () => void }): JSX.Element => {
    const [loading, setLoading] = useState(false);
    const [fetchError, setFetchError] = useState(false);
    const actionFunc = useCallback(async () => {
        setLoading(true);
        try {
            const response = await fetch(`${HTTP_URL}/remove`, {
                method: "POST",
                body: JSON.stringify({
                    id: id,
                }),
            });
            if (response?.status !== 204) {
                throw Error(await response.text());
            }
        } catch (e) {
            console.log(e);
            setLoading(false);
            setFetchError(true);
            return;
        }

        closeFunc()
    }, [id, closeFunc]);

    return <Dialog open onClose={closeFunc}>
        <DialogTitle>
            Are you want to remove this item?
        </DialogTitle>
        {(loading || fetchError) && <DialogContent dividers>
            {loading && <Alert icon={<CircularProgress/>} severity="info">
                Removing...
            </Alert>}

            {fetchError && <Alert icon={<ErrorIcon/>} severity="error">
                Error while removing, please check console
            </Alert>}
        </DialogContent>}
        <DialogActions>
            <Button onClick={closeFunc}>Cancel</Button>
            <Button onClick={actionFunc} autoFocus>Yes</Button>
        </DialogActions>
    </Dialog>;
}

export const RemoveAccessDialog = ({id, closeFunc}: { id: string, closeFunc: () => void }): JSX.Element => {
    const [loading, setLoading] = useState(false);
    const [fetchError, setFetchError] = useState(false);
    const actionFunc = useCallback(async () => {
        setLoading(true);
        try {
            const response = await fetch(`${HTTP_URL}/remove-access`, {
                method: "POST",
                body: JSON.stringify({
                    id: id,
                }),
            });
            if (response?.status !== 204) {
                throw Error(await response.text());
            }
        } catch (e) {
            console.log(e);
            setLoading(false);
            setFetchError(true);
            return;
        }

        closeFunc()
    }, [id, closeFunc]);

    return <Dialog open onClose={closeFunc}>
        <DialogTitle>
            Are you want to remove the access for this item?
        </DialogTitle>
        {(loading || fetchError) && <DialogContent dividers>
            Are you want to remove the access for this item?

            {loading && <Alert icon={<CircularProgress/>} severity="info">
                Removing access...
            </Alert>}

            {fetchError && <Alert icon={<ErrorIcon/>} severity="error">
                Error while removing access, please check console
            </Alert>}
        </DialogContent>}
        <DialogActions>
            <Button onClick={closeFunc}>Cancel</Button>
            <Button onClick={actionFunc} autoFocus>Yes</Button>
        </DialogActions>
    </Dialog>;
}
