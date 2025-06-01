import React, { useEffect, useState, useRef } from 'react';
import {
  Banner,
  Button,
  Col,
  Form,
  Row,
  Spin,
  Collapse,
  Modal,
  TagInput,
} from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function GeneralSettings(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [showQuotaWarning, setShowQuotaWarning] = useState(false);
  const [inputs, setInputs] = useState({
    TopUpLink: '',
    'general_setting.docs_link': '',
    QuotaPerUnit: '',
    RetryTimes: '',
    DisplayInCurrencyEnabled: false,
    DisplayTokenStatEnabled: false,
    DefaultCollapseSidebar: false,
    DemoSiteEnabled: false,
    SelfUseModeEnabled: false,
    CheckinEnabled: false,
    BaseCheckinReward: 10000,
    ContinuousCheckinReward: 1000,
    MaxContinuousRewardDays: 7,
    CheckinStreakReset: true,
    SpecialRewardDays: [7, 15, 30],
    SpecialRewards: [20000, 50000, 100000],
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function handleFieldChange(fieldName) {
    return (value) => {
      setInputs((inputs) => ({ ...(inputs || {}), [fieldName]: value }));
    };
  }

  function onSubmit() {
    // 创建一个更新项目的数组
    let updateArray = [];
    
    // 遍历所有输入项，检查是否有变化
    for (const key in inputs) {
      if (inputs[key] !== inputsRow[key]) {
        updateArray.push({
          key: key,
          oldValue: inputsRow[key],
          newValue: inputs[key]
        });
      }
    }
    
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    
    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof item.newValue === 'boolean') {
        value = String(item.newValue);
      } else if (Array.isArray(item.newValue) && (item.key === 'SpecialRewardDays' || item.key === 'SpecialRewards')) {
        // 对数组类型的特殊选项进行JSON字符串转换
        value = JSON.stringify(item.newValue);
      } else {
        value = item.newValue;
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Banner
          type='warning'
          description={t('聊天链接功能已经弃用，请使用下方聊天设置功能')}
        />
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('通用设置')}>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Input
                  field={'TopUpLink'}
                  label={t('充值链接')}
                  initValue={''}
                  placeholder={t('例如发卡网站的购买链接')}
                  onChange={handleFieldChange('TopUpLink')}
                  showClear
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Input
                  field={'general_setting.docs_link'}
                  label={t('文档地址')}
                  initValue={''}
                  placeholder={t('例如 https://docs.newapi.pro')}
                  onChange={handleFieldChange('general_setting.docs_link')}
                  showClear
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Input
                  field={'QuotaPerUnit'}
                  label={t('单位美元额度')}
                  initValue={''}
                  placeholder={t('一单位货币能兑换的额度')}
                  onChange={handleFieldChange('QuotaPerUnit')}
                  showClear
                  onClick={() => setShowQuotaWarning(true)}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Input
                  field={'RetryTimes'}
                  label={t('失败重试次数')}
                  initValue={''}
                  placeholder={t('失败重试次数')}
                  onChange={handleFieldChange('RetryTimes')}
                  showClear
                />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'DisplayInCurrencyEnabled'}
                  label={t('以货币形式显示额度')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={handleFieldChange('DisplayInCurrencyEnabled')}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'DisplayTokenStatEnabled'}
                  label={t('额度查询接口返回令牌额度而非用户额度')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={handleFieldChange('DisplayTokenStatEnabled')}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'DefaultCollapseSidebar'}
                  label={t('默认折叠侧边栏')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={handleFieldChange('DefaultCollapseSidebar')}
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'DemoSiteEnabled'}
                  label={t('演示站点模式')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={handleFieldChange('DemoSiteEnabled')}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'SelfUseModeEnabled'}
                  label={t('自用模式')}
                  extraText={t('开启后不限制：必须设置模型倍率')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={handleFieldChange('SelfUseModeEnabled')}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'CheckinEnabled'}
                  label={t('启用签到功能')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={handleFieldChange('CheckinEnabled')}
                />
              </Col>
            </Row>
            
            {/* 签到设置区域 */}
            {inputs.CheckinEnabled && (
              <Row gutter={16}>
                <Col xs={24}>
                  <Form.Section text={t('签到奖励设置')}>
                    <Row gutter={16}>
                      <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                        <Form.InputNumber
                          field={'BaseCheckinReward'}
                          label={t('基础签到奖励')}
                          initValue={10000}
                          min={0}
                          placeholder={t('每次签到获得的基础奖励')}
                          onChange={handleFieldChange('BaseCheckinReward')}
                        />
                      </Col>
                      <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                        <Form.InputNumber
                          field={'ContinuousCheckinReward'}
                          label={t('连续签到额外奖励')}
                          initValue={1000}
                          min={0}
                          placeholder={t('连续签到每天额外奖励')}
                          onChange={handleFieldChange('ContinuousCheckinReward')}
                        />
                      </Col>
                      <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                        <Form.InputNumber
                          field={'MaxContinuousRewardDays'}
                          label={t('最大连续奖励天数')}
                          initValue={7}
                          min={1}
                          max={30}
                          placeholder={t('连续签到奖励最多累计天数')}
                          onChange={handleFieldChange('MaxContinuousRewardDays')}
                        />
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                        <Form.Switch
                          field={'CheckinStreakReset'}
                          label={t('中断签到重置连续计数')}
                          initValue={true}
                          size='default'
                          checkedText='｜'
                          uncheckedText='〇'
                          onChange={handleFieldChange('CheckinStreakReset')}
                        />
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col xs={24} sm={12}>
                        <Form.TagInput
                          field={'SpecialRewardDays'}
                          label={t('特殊奖励天数')}
                          initValue={[7, 15, 30]}
                          placeholder={t('输入天数后按回车，如: 7, 15, 30')}
                          validateStatus='default'
                          showClear
                          onChange={(value) => {
                            // 确保输入的是数字
                            const numericValues = value
                              .map((v) => parseInt(v))
                              .filter((v) => !isNaN(v) && v > 0);
                            handleFieldChange('SpecialRewardDays')(numericValues);
                          }}
                          onInputExceed={() => showWarning(t('已达到最大输入数量'))}
                          addOnBlur={true}
                        />
                      </Col>
                      <Col xs={24} sm={12}>
                        <Form.TagInput
                          field={'SpecialRewards'}
                          label={t('特殊天数奖励额度')}
                          initValue={[20000, 50000, 100000]}
                          placeholder={t('输入奖励额度后按回车，如: 20000')}
                          validateStatus='default'
                          showClear
                          onChange={(value) => {
                            // 确保输入的是数字
                            const numericValues = value
                              .map((v) => parseInt(v))
                              .filter((v) => !isNaN(v) && v > 0);
                            handleFieldChange('SpecialRewards')(numericValues);
                          }}
                          onInputExceed={() => showWarning(t('已达到最大输入数量'))}
                          addOnBlur={true}
                        />
                      </Col>
                    </Row>
                    <Row>
                      <Banner
                        type='info'
                        description={t('特殊奖励天数和奖励额度需一一对应，例如第7天奖励20000，第15天奖励50000，第30天奖励100000')}
                        bordered
                        fullMode={false}
                      />
                    </Row>
                  </Form.Section>
                </Col>
              </Row>
            )}
            
            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存通用设置')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>

      <Modal
        title={t('警告')}
        visible={showQuotaWarning}
        onOk={() => setShowQuotaWarning(false)}
        onCancel={() => setShowQuotaWarning(false)}
        closeOnEsc={true}
        width={500}
      >
        <Banner
          type='warning'
          description={t(
            '此设置用于系统内部计算，默认值500000是为了精确到6位小数点设计，不推荐修改。',
          )}
          bordered
          fullMode={false}
          closeIcon={null}
        />
      </Modal>
    </>
  );
}
