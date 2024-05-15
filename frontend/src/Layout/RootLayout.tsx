import './Layout.css';
import {Outlet} from 'react-router-dom';
import {Alerts} from "../Context/Alert.tsx";
import {useTheme} from "@mui/material";
import {CSSProperties} from "react";

const style = {
//    backgroundImage: "url('/background.png')",
//    backgroundSize: 'cover',
    backgroundColor: 'rgb(40,40,40)'
} as CSSProperties;

export const RootLayout = () => {
    const theme = useTheme()
    return (
        <>
            <Alerts/>
            <div className={'layout'} style={{...style, backgroundColor: theme.palette.background.default}}>
                <Outlet/>
            </div>
        </>
    )
}
