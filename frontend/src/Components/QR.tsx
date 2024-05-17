import {QRCode} from "react-qrcode-logo";
import {useMemo} from "react";
import {useTheme} from "@mui/material";

export type QRProps = {
    value?: string
}

export const QR = ({value}: QRProps) => {
    const theme = useTheme();
    const logoSettings = useMemo(() => {
        return {
            logoImage: '',
            eyeRadius: 0,
            logoPadding: 0,
            logoPaddingStyle: 'circle' as ('circle' | 'square' | undefined)
        }
    }, [])
    return (
    <QRCode bgColor={theme.palette.background.paper} fgColor={theme.palette.text.primary} size={300}
            ecLevel={'L'}
            value={value ? value : `${location.toString()}`} qrStyle={'squares'} {...logoSettings} />
    );
}
