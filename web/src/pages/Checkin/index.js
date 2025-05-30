import React, { useState } from 'react';
import { Button, Card } from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../helpers';
import { useTranslation } from 'react-i18next';

const Checkin = () => {
  const { t } = useTranslation();
  const [days, setDays] = useState(null);
  const [done, setDone] = useState(false);

  const handleCheckin = async () => {
    try {
      const res = await API.post('/api/checkin');
      const { success, message, data } = res.data;
      if (success) {
        setDays(data.continuous);
        setDone(true);
        showSuccess(t('签到成功'));
      } else {
        showError(message || t('签到失败'));
      }
    } catch (err) {
      showError(t('签到失败'));
    }
  };

  return (
    <Card style={{ maxWidth: 400, margin: '0 auto' }} title={t('每日签到')}>
      <p>{days ? t('已连续签到 {{days}} 天', { days }) : t('快来签到吧！')}</p>
      <Button theme='solid' onClick={handleCheckin} disabled={done}>
        {done ? t('已签到') : t('签到')}
      </Button>
    </Card>
  );
};

export default Checkin;
