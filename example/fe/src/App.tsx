import {useEffect, useState} from 'react'
import './App.css'
import {LiveDataError, LiveDataLoading, LiveDataReady, useLiveData} from "./useLiveData.ts";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import TableBody from "@mui/material/TableBody";
import TableContainer from "@mui/material/TableContainer";
import {SMap} from "./utils/SMap.ts";
import {AddDialog} from "./components/AddDialog.tsx";
import Button from "@mui/material/Button";
import AddIcon from '@mui/icons-material/Add';
import {EditDialog, EditDialogBag} from "./components/EditDialog.tsx";
import {IconButton} from "@mui/material";
import DeleteIcon from '@mui/icons-material/Delete';
import BlockIcon from '@mui/icons-material/Block';
import EditIcon from '@mui/icons-material/Edit';
import {RemoveAccessDialog, RemoveDialog} from "./components/RemoveDialogs.tsx";
import {nanoid} from "nanoid";
import {getLiveData} from "./rpr/RPR.ts";
import {proxy} from "comlink";
import {OFFLINE_STATUS, ONLINE_STATUS} from "./rpr/RPRConst.ts";

export const App = () => {
    const {liveDataState, liveData} = useLiveData("kv");

    useEffect(() => {
        console.log(liveData, liveDataState);
    }, [liveData, liveDataState]);

    const [connectionStatus, setConnectionStatus] = useState<string>();

    useEffect(() => {
        const id = nanoid();
        getLiveData().subscribeConnectionStatus(id, proxy(setConnectionStatus));
        return () => {
            getLiveData().unsubscribeConnectionStatus(id);
        }
    }, []);

    const [addDialogShown, setAddDialogShown] = useState(false);
    const [editDialogBag, setEditDialogBag] = useState<EditDialogBag>();
    const [removeDialogId, setRemoveDialogId] = useState<string>();
    const [removeAccessDialogId, setRemoveAccessDialogId] = useState<string>();

    return <>
        <h1>Realtime Persistent Replication demo -
            {connectionStatus == ONLINE_STATUS && <i style={{color: "green"}}> {connectionStatus}</i>}
            {connectionStatus == OFFLINE_STATUS && <i style={{color: "red"}}> {connectionStatus}</i>}
        </h1>
        <h6>Automatically reconnects every 15 second when connection has been lost</h6>

        <TableContainer component={Paper}>
            <Table sx={{minWidth: 650}}>
                <TableHead>
                    <TableRow>
                        <TableCell>Key</TableCell>
                        <TableCell align="right">Value</TableCell>
                        <TableCell align="right">
                            <Button variant="outlined" startIcon={<AddIcon/>} onClick={() => setAddDialogShown(true)}>
                                Add
                            </Button>
                        </TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {liveDataState === LiveDataLoading && <TableRow style={{height: 200}}>
                        <TableCell colSpan={5}>
                            Loading...
                        </TableCell>
                    </TableRow>}

                    {liveDataState === LiveDataError && <TableRow style={{height: 200}}>
                        <TableCell colSpan={5}>
                            Error while loading, check console
                        </TableCell>
                    </TableRow>}

                    {liveDataState === LiveDataReady && <>
                        {SMap.foreach(liveData).map(([id, data]) => (
                            <TableRow key={id} sx={{'&:last-child td, &:last-child th': {border: 0}}}>
                                <TableCell component="th" scope="row">
                                    {id}
                                </TableCell>
                                <TableCell align="right">{data?.value as string}</TableCell>
                                <TableCell align="right">
                                    <IconButton size="small" onClick={() => setEditDialogBag({
                                        id: id,
                                        value: data?.value as string,
                                    })}>
                                        <EditIcon fontSize="small"/>
                                    </IconButton>
                                    <IconButton size="small" onClick={() => setRemoveDialogId(id)}>
                                        <DeleteIcon fontSize="small"/>
                                    </IconButton>
                                    <IconButton size="small" onClick={() => setRemoveAccessDialogId(id)}>
                                        <BlockIcon fontSize="small"/>
                                    </IconButton>
                                </TableCell>
                            </TableRow>
                        ))}
                    </>}
                </TableBody>
            </Table>
        </TableContainer>

        {addDialogShown && <AddDialog closeFunc={() => setAddDialogShown(false)}/>}
        {editDialogBag != null && <EditDialog bag={editDialogBag} closeFunc={() => setEditDialogBag(null)}/>}
        {removeDialogId != null && <RemoveDialog id={removeDialogId} closeFunc={() => setRemoveDialogId(null)}/>}
        {removeAccessDialogId != null &&
            <RemoveAccessDialog id={removeAccessDialogId} closeFunc={() => setRemoveAccessDialogId(null)}/>}
    </>;
};
