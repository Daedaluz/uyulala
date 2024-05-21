import './Layout.css';
import {Outlet} from 'react-router-dom';
import {Alerts} from "../Context/Alert.tsx";
import {useTheme} from "@mui/material";
import {useMemo} from "react";

export const RootLayout = () => {
    const theme = useTheme()
    const bg = useMemo(()=> {
        const bgArray = ["1.jpg", "2.jpg"];
        const index = Math.floor(Math.random() * bgArray.length);
        return `url('/bg/${bgArray[index]}')`;
    }, []);

    return (
        <>
            <Alerts/>
            <div className={'layout'} style={{backgroundColor: theme.palette.background.default, backgroundImage: bg}}>
                <Outlet/>
            </div>
        </>
    )
}
