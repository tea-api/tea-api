import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  API,
  downloadTextAsFile,
  isMobile,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';
import {
  getQuotaPerUnit,
  renderQuota,
  renderQuotaWithPrompt,
} from '../../helpers/render';
import {
  AutoComplete,
  Button,
  DatePicker,
  Input,
  Modal,
  SideSheet,
  Space,
  Spin,
  Typography,
} from '@douyinfe/semi-ui';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { Divider } from 'semantic-ui-react';

const EditRedemption = (props) => {
  const { t } = useTranslation();
  const isEdit = props.editingRedemption.id !== undefined;
  const [loading, setLoading] = useState(isEdit);

  const params = useParams();
  const navigate = useNavigate();
  const originInputs = {
    name: '',
    quota: 100000,
    count: 1,
    key: '',
    max_times: 1,
    expired_time: -1,
  };
  const [inputs, setInputs] = useState(originInputs);
  const { name, quota, count, key, max_times, expired_time } = inputs;

  const handleCancel = () => {
    props.handleClose();
  };

  const setExpiredTime = (month, day, hour, minute) => {
    let now = new Date();
    let timestamp = now.getTime() / 1000;
    let seconds = month * 30 * 24 * 60 * 60;
    seconds += day * 24 * 60 * 60;
    seconds += hour * 60 * 60;
    seconds += minute * 60;
    if (seconds !== 0) {
      timestamp += seconds;
      setInputs({ ...inputs, expired_time: timestamp2string(timestamp) });
    } else {
      setInputs({ ...inputs, expired_time: -1 });
    }
  };

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadRedemption = async () => {
    setLoading(true);
    let res = await API.get(`/api/redemption/${props.editingRedemption.id}`);
    const { success, message, data } = res.data;
    if (success) {
      if (data.expired_time !== -1) {
        data.expired_time = timestamp2string(data.expired_time);
      }
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    if (isEdit) {
      loadRedemption().then(() => {
        // console.log(inputs);
      });
    } else {
      setInputs(originInputs);
    }
  }, [props.editingRedemption.id]);

  const submit = async () => {
    let name = inputs.name;
    if (!isEdit && inputs.name === '') {
      // set default name
      name = renderQuota(quota);
    }
    setLoading(true);
    let localInputs = inputs;
    localInputs.count = parseInt(localInputs.count);
    localInputs.quota = parseInt(localInputs.quota);
    localInputs.max_times = parseInt(localInputs.max_times);
    localInputs.name = name;
    if (localInputs.expired_time !== -1) {
      let time = Date.parse(localInputs.expired_time);
      if (isNaN(time)) {
        showError(t('过期时间格式错误！'));
        setLoading(false);
        return;
      }
      localInputs.expired_time = Math.ceil(time / 1000);
    }
    let res;
    if (isEdit) {
      res = await API.put(`/api/redemption/`, {
        ...localInputs,
        id: parseInt(props.editingRedemption.id),
      });
    } else {
      res = await API.post(`/api/redemption/`, {
        ...localInputs,
      });
    }
    const { success, message, data } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess(t('兑换码更新成功！'));
        props.refresh();
        props.handleClose();
      } else {
        showSuccess(t('兑换码创建成功！'));
        setInputs(originInputs);
        props.refresh();
        props.handleClose();
      }
    } else {
      showError(message);
    }
    if (!isEdit && data) {
      let text = '';
      for (let i = 0; i < data.length; i++) {
        text += data[i] + '\n';
      }
      Modal.confirm({
        title: t('兑换码创建成功'),
        content: (
          <div>
            <p>{t('兑换码创建成功，是否下载兑换码？')}</p>
            <p>{t('兑换码将以文本文件的形式下载，文件名为兑换码的名称。')}</p>
          </div>
        ),
        onOk: () => {
          downloadTextAsFile(text, `${inputs.name}.txt`);
        },
      });
    }
    setLoading(false);
  };

  return (
    <>
      <SideSheet
        placement={isEdit ? 'right' : 'left'}
        title={
          <Title level={3}>
            {isEdit ? t('更新兑换码信息') : t('创建新的兑换码')}
          </Title>
        }
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visiable}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme='solid' size={'large'} onClick={submit}>
                {t('提交')}
              </Button>
              <Button
                theme='solid'
                size={'large'}
                type={'tertiary'}
                onClick={handleCancel}
              >
                {t('取消')}
              </Button>
            </Space>
          </div>
        }
        closeIcon={null}
        onCancel={() => handleCancel()}
        width={isMobile() ? '100%' : 600}
      >
        <Spin spinning={loading}>
          <Input
            style={{ marginTop: 20 }}
            label={t('名称')}
            name='name'
            placeholder={t('请输入名称')}
            onChange={(value) => handleInputChange('name', value)}
            value={name}
            autoComplete='new-password'
            required={!isEdit}
          />
          {!isEdit && (
            <>
              <Divider />
              <Input
                style={{ marginTop: 8 }}
                label={t('自定义兑换码')}
                name='key'
                placeholder={t('自定义兑换码')}
                onChange={(value) => handleInputChange('key', value)}
                value={key}
                autoComplete='new-password'
              />
              <Typography.Text type="tertiary" style={{ marginTop: 4, display: 'block' }}>
                {t('可以输入完整兑换码（仅限生成1个）或输入前缀+"*"批量生成（如"GIFT2024*"）')}
              </Typography.Text>
            </>
          )}
          <Divider />
          <div style={{ marginTop: 20 }}>
            <Typography.Text>
              {t('额度') + renderQuotaWithPrompt(quota)}
            </Typography.Text>
          </div>
          <AutoComplete
            style={{ marginTop: 8 }}
            name='quota'
            placeholder={t('请输入额度')}
            onChange={(value) => handleInputChange('quota', value)}
            value={quota}
            autoComplete='new-password'
            type='number'
            position={'bottom'}
            data={[
              { value: 500000, label: '1$' },
              { value: 5000000, label: '10$' },
              { value: 25000000, label: '50$' },
              { value: 50000000, label: '100$' },
              { value: 250000000, label: '500$' },
              { value: 500000000, label: '1000$' },
            ]}
          />
          <Divider />
          <div style={{ marginTop: 20 }}>
            <Typography.Text>
              {t('最多可用次数') + '（每个兑换码可被使用的最大次数）'}
            </Typography.Text>
          </div>
          <Input
            style={{ marginTop: 8 }}
            label={t('最多使用次数')}
            name='max_times'
            placeholder={t('最多使用次数')}
            onChange={(value) => handleInputChange('max_times', value)}
            value={max_times}
            autoComplete='new-password'
            type='number'
          />
          <Divider />
          <DatePicker
            label={t('过期时间')}
            name='expired_time'
            placeholder={t('请选择过期时间')}
            onChange={(value) => handleInputChange('expired_time', value)}
            value={expired_time}
            autoComplete='new-password'
            type='dateTime'
          />
          <div style={{ marginTop: 8 }}>
            <Space>
              <Button type={'tertiary'} onClick={() => setExpiredTime(0, 0, 0, 0)}>
                {t('永不过期')}
              </Button>
              <Button type={'tertiary'} onClick={() => setExpiredTime(0, 0, 1, 0)}>
                {t('一小时')}
              </Button>
              <Button type={'tertiary'} onClick={() => setExpiredTime(0, 1, 0, 0)}>
                {t('一天')}
              </Button>
              <Button type={'tertiary'} onClick={() => setExpiredTime(1, 0, 0, 0)}>
                {t('一个月')}
              </Button>
            </Space>
          </div>
          {!isEdit && (
            <>
              <Divider />
              <Typography.Text>{t('生成数量')}</Typography.Text>
              <Input
                style={{ marginTop: 8 }}
                label={t('生成数量')}
                name='count'
                placeholder={t('请输入生成数量')}
                onChange={(value) => handleInputChange('count', value)}
                value={count}
                autoComplete='new-password'
                type='number'
              />
            </>
          )}
        </Spin>
      </SideSheet>
    </>
  );
};

export default EditRedemption;
