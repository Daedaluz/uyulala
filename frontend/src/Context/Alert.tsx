import {ReactNode, createContext, useCallback, useContext, useMemo, useRef, useState} from "react";
import {Alert, AlertTitle, useMediaQuery, useTheme} from "@mui/material";
import CSS from 'csstype';

export type AlertEntry = {
    id: number
    severity: 'error' | 'warning' | 'info' | 'success'
    title?: string
    message: string
    timeout?: number
    onClose?: () => void
}

export type AlertContext = {
    alert: AlertEntry[]
    showAlert: (severity: string, title: string, message: string, timeout?: number) => void
}

const alertContext = createContext({alert: [], showAlert: () => {}} as AlertContext);

type AlertProviderProps = {
    children: ReactNode
}

export const AlertProvider = ({children}: AlertProviderProps) => {
    const [alert, setAlert] = useState<AlertEntry[]>([]);
    const id = useRef(0)
    const showAlert = useCallback((severity: string, title: string, message: string, timeout?: number) => {
        const aid = id.current++;
        const alertEntry = {
            id: aid, severity, title, message, timeout
        } as AlertEntry;
        if (timeout === undefined) {
            alertEntry.onClose = () => {
                setAlert((prev) => prev.filter((a) => a.id !== aid));
            }
        }
        setAlert((prev) => [alertEntry, ...prev]);
        if (timeout) {
            setTimeout(() => {
                setAlert((prev) => prev.filter((a) => a.id !== aid));
            }, timeout);
        }
    }, [setAlert]);

    const contextValue = useMemo(() => ({alert, showAlert}), [alert, showAlert]);
    return (
        <alertContext.Provider value={contextValue}>
            {children}
        </alertContext.Provider>
    )
};


const alertStyle = {
    position: 'absolute',
    alignSelf: 'right',
    marginTop: '1em',
    marginRight: '1em',
    marginLeft: '1em',
    top: 0,
    right: 0,
    bottom: 0,
} as CSS.Properties;

export const Alerts = () => {
    const theme = useTheme();
    const {alert} = useAlert();
    const md = useMediaQuery(theme.breakpoints.up('md'));
    const maxWidth = md ? '30vw' : '100vw';
    return (
        <div style={{...alertStyle, maxWidth}}>
            {alert.map((a) => {
                return (
                    <Alert variant={'filled'} key={a.id} severity={a.severity} onClose={a.onClose}>
                        {a.title && <AlertTitle>{a.title}</AlertTitle>}
                        {a.message}
                    </Alert>
                )
            })}
        </div>
    );
};

export const useAlert = () => {
    const context = useContext(alertContext);
    if (context === undefined) {
        throw new Error('useAlert must be used within a AlertProvider')
    }
    return context;
};
