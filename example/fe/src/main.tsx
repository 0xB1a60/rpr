import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import {App} from "./App.tsx";
import {initRPR} from "./rpr/RPR.ts";

ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>,
)

initRPR();
