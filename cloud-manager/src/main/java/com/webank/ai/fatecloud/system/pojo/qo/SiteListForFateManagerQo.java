package com.webank.ai.fatecloud.system.pojo.qo;

import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.*;

import java.io.Serializable;

@Data
@NoArgsConstructor
@AllArgsConstructor
@ToString
@EqualsAndHashCode
@ApiModel("find for site list for fate manager")
public class SiteListForFateManagerQo implements Serializable {

    @ApiModelProperty("pageNum")
    private Integer pageNum = 1;

    @ApiModelProperty("pageSize")
    private Integer pageSize = 10;

    @ApiModelProperty(value = "institutions")
    private String institutions;
}
