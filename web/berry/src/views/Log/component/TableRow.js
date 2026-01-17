import PropTypes from 'prop-types';

import { TableRow, TableCell, IconButton } from '@mui/material';

import { timestamp2string, renderQuota } from 'utils/common';
import Label from 'ui-component/Label';
import LogType from '../type/LogType';
import { IconEye } from '@tabler/icons-react';

function renderType(type) {
  const typeOption = LogType[type];
  if (typeOption) {
    return (
      <Label variant="filled" color={typeOption.color}>
        {' '}
        {typeOption.text}{' '}
      </Label>
    );
  } else {
    return (
      <Label variant="filled" color="error">
        {' '}
        未知{' '}
      </Label>
    );
  }
}

export default function LogTableRow({ item, userIsAdmin, onShowDetail }) {
  return (
    <>
      <TableRow tabIndex={item.id} hover>
        <TableCell>{timestamp2string(item.created_at)}</TableCell>

        {userIsAdmin && <TableCell>{item.channel || ''}</TableCell>}
        {userIsAdmin && (
          <TableCell>
            <Label color="default" variant="outlined">
              {item.username}
            </Label>
          </TableCell>
        )}
        <TableCell>
          {item.token_name && (
            <Label color="default" variant="soft">
              {item.token_name}
            </Label>
          )}
        </TableCell>
        <TableCell>{renderType(item.type)}</TableCell>
        <TableCell>
          {item.model_name && (
            <Label color="primary" variant="outlined">
              {item.model_name}
            </Label>
          )}
        </TableCell>
          <TableCell>{`${Number(item.prompt_tokens ?? 0).toLocaleString('en-US')} → ${Number(item.completion_tokens ?? 0).toLocaleString('en-US')}`}</TableCell>

        <TableCell>{item.first_token_time || '0'}</TableCell>
        <TableCell>
          {Number(item.completion_tokens) > 0 && Number(item.elapsed_time) > 0
            ? `${Math.round(((Number(item.completion_tokens) / Number(item.elapsed_time)) * 1000) * 10) / 10} T/s`
            : ''}
        </TableCell>

        <TableCell>{item.quota ? renderQuota(item.quota, 6) : ''}</TableCell>
        <TableCell>
          <IconButton
            size="small"
            onClick={() => onShowDetail(item)}
            color="primary"
            title="查看详情"
          >
            <IconEye size={18} />
          </IconButton>
        </TableCell>
      </TableRow>
    </>
  );
}

LogTableRow.propTypes = {
  item: PropTypes.object,
  userIsAdmin: PropTypes.bool,
  onShowDetail: PropTypes.func
};
