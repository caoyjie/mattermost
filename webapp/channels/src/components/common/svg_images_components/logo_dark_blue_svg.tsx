// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import styled from 'styled-components';

import logo from 'images/logo.svg';

type Props = {
    width?: number;
    height?: number;
    className?: string;
}

const LogoImage = styled.img.attrs({
    alt: 'Mattermost logo',
})``;

export default (props: Props) => (
    <LogoImage
        className={props.className}
        width={props.width ? props.width.toString() : '182'}
        height={props.height ? props.height.toString() : '30'}
        src={logo}
    />
);
