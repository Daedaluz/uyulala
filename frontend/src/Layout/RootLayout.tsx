import './Layout.css';
import {Outlet} from 'react-router-dom';
import {Alerts} from "../Context/Alert.tsx";
import {useTheme} from "@mui/material";

export const RootLayout = () => {
    const theme = useTheme()
    return (
        <>
            <Alerts/>
            <div className={'layout'} style={{backgroundColor: theme.palette.background.default}}>
                <Outlet/>
            </div>
        </>
    )
}