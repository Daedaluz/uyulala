import {App} from "../Api/public.ts";
import {QRCode} from "react-qrcode-logo";
import {useMemo} from "react";

export type QRProps = {
    app: App
}

export const QR = ({app}: QRProps) => {
    const logoSettings = useMemo(() => {
        if (app.icon !== '') {
            return {
                logoImage: app.icon,
                eyeRadius: 7,
                logoPadding: 5,
                logoPaddingStyle: 'circle' as ('circle' | 'square' | undefined)
            }
        }
        return {
            logoImage: '',
            eyeRadius: 0,
            logoPadding: 0,
            logoPaddingStyle: 'circle' as ('circle' | 'square' | undefined)
        }
    }, [app.icon])
    return (
        <QRCode bgColor={'#FFFFFF00'} fgColor={"#FFFFFF"} removeQrCodeBehindLogo={true} size={300}
                ecLevel={'H'}
                value={`${location.toString()}`} qrStyle={'dots'} {...logoSettings} />
    );
}